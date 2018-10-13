package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkInfo;
import android.net.NetworkRequest;
import android.os.Bundle;
import android.telephony.TelephonyManager;
import android.util.Base64;
import android.util.Log;

import com.android.mms.transaction.HttpUtils;
import com.android.mms.transaction.TransactionSettings;
import com.android.volley.Request;
import com.android.volley.RequestQueue;
import com.android.volley.Response;
import com.android.volley.VolleyError;
import com.android.volley.toolbox.StringRequest;
import com.android.volley.toolbox.Volley;
import com.google.android.mms_clone.ContentType;
import com.google.android.mms_clone.pdu.EncodedStringValue;
import com.google.android.mms_clone.pdu.GenericPdu;
import com.google.android.mms_clone.pdu.MultimediaMessagePdu;
import com.google.android.mms_clone.pdu.NotificationInd;
import com.google.android.mms_clone.pdu.PduBody;
import com.google.android.mms_clone.pdu.PduHeaders;
import com.google.android.mms_clone.pdu.PduParser;
import com.google.android.mms_clone.pdu.PduPart;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Consumer;

/**
 * Receives SMS messages from the system to send upstream.
 */
public class MMSReceiver extends BroadcastReceiver {
	//Holds the different types of headers that represent the "to" field.
	private static final int[] TO_ADDRESS_TYPES = {PduHeaders.TO, PduHeaders.BCC, PduHeaders.CC};
	private static final String RECIPIENTS_KEY = "recipients";
	private static final String BLOCK_ID_KEY = "block_id";
	private static final String ILLEGAL_NO_HOST_MESSAGE = "No host has been set";
	private static final String ILLEGAL_NO_ID_MESSAGE = "No device id has been set";
	private static final String ILLEGAL_NO_SESSION_MESSAGE = "No session id has been set";

	/**
	 * Sends different MMS parts upstream
	 */
	private class PartSender {
		private AtomicInteger numSent;
		private List<String> partsList;
		private String blockId = null;
		private Consumer<String> callback;
		private RequestQueue reqQueue;

		/**
		 * Construct a PartSender
		 * @param context The context in which the receiver is running
		 * @param partsList A list of the parts as b64 encoded strings.
		 * @param callback A callback that takes the block id once the parts have been sent upstream.
		 */
		private PartSender(Context context, List<String> partsList, Consumer<String> callback) {
			this.partsList = partsList;
			this.callback = callback;
			this.numSent = new AtomicInteger();
			this.reqQueue = Volley.newRequestQueue(context);
		}

		/**
		 * Send the MMS parts upstream.
		 * @param context The context in which the receiver is running.
		 */
		private void send(Context context) {
			SharedPreferences prefs = context.getSharedPreferences(MainActivity.PREFS_KEY, Context.MODE_PRIVATE);
			String host = prefs.getString(MainActivity.HOST_URL_PREFS_KEY, "");
			String sessionId = prefs.getString(MainActivity.SESSION_ID_PREFS_KEY, "");
			final String deviceId = prefs.getString(MainActivity.DEVICE_ID_PREFS_KEY, "");
			if (host.length() == 0) {
				Log.wtf("SMSPusher", ILLEGAL_NO_HOST_MESSAGE);
			}
			if (deviceId.length() == 0) {
				Log.wtf("SMSPusher", ILLEGAL_NO_ID_MESSAGE);
			}
			if (sessionId.length() == 0) {
				Log.wtf("SMSPusher", ILLEGAL_NO_SESSION_MESSAGE);
			}
			try {
				final URL hostUrl = new URL(host);
				final URL uploadUrl = new URL(hostUrl, "/upload_mms_file");
				final String firstPart = partsList.remove(0);
				final Response.Listener<String> respListener = (String response) -> {
					try {
						JSONObject resObject = new JSONObject(response);
						blockId = (String) resObject.get("block_id");
						int numSentSoFar = numSent.incrementAndGet();
						//If we have all of the parts, call our callback.
						if (numSentSoFar >= partsList.size()) {
							callback.accept(blockId);
						}
					} catch (JSONException e) {
						Log.e("SMSPusher", e.toString());
					}
				};

				//A consistent error listener for all requests, simply logs the error for us.
				final Response.ErrorListener errorListener = (VolleyError e) -> Log.e("SMSPusher", e.toString());

				//Make a request that will add all the others once it is done - this allows for us to have the block-id set.
				StringRequest req = new StringRequest(Request.Method.POST, uploadUrl.toString(), (String response) -> {
					respListener.onResponse(response);
					for (final String part : partsList) {
						//subsequentReq is a request that will include the current block id of the mms parts.
						StringRequest subsequentReq = new StringRequest(Request.Method.POST, uploadUrl.toString(), respListener, errorListener) {
							protected Map<String, String> getParams() {
								Map<String, String> paramsMap = new HashMap<>();
								paramsMap.put("device_id", deviceId);
								paramsMap.put("data", part);
								paramsMap.put("session_id", sessionId);
								paramsMap.put("block_id", blockId);
								return paramsMap;
							}
						};
						reqQueue.add(subsequentReq);
					}
				}, errorListener) {
					protected Map<String, String> getParams() {
						Map<String, String> paramsMap = new HashMap<>();
						paramsMap.put("device_id", deviceId);
						paramsMap.put("data", firstPart);
						paramsMap.put("session_id", sessionId);
						return paramsMap;
					}
				};

				reqQueue.add(req);
			} catch (MalformedURLException e) {
				Log.wtf("SMSPusher", e);
				e.printStackTrace();
			}
		}
	}

	/**
	 * Handles the receipt of an MMS message, sending it upstream
	 * @param context The context in which the receiver is running
	 * @param intent The intent being received.
	 */
	@Override
	public void onReceive(final Context context, Intent intent) {
		Bundle extras = intent.getExtras();
		byte[] data = extras.getByteArray("data");
		PduParser parser = new PduParser(data);
		GenericPdu genericPdu = parser.parse();
		if (genericPdu.getMessageType() == PduHeaders.MESSAGE_TYPE_NOTIFICATION_IND) {
			NotificationInd pdu = (NotificationInd) genericPdu;
			//Get the location of the MMS content, and get the message content.
			byte[] rawContentLocation = pdu.getContentLocation();
			final String contentLocation = new String(rawContentLocation);
			final ConnectivityManager connectivityManager = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
			getNetwork(connectivityManager, new ConnectivityManager.NetworkCallback() {
				@Override
				public void onAvailable(Network network) {
					connectivityManager.bindProcessToNetwork(network);
					NetworkInfo info = connectivityManager.getNetworkInfo(network);
					String extraInfo = info.getExtraInfo();
					TransactionSettings transactionSettings = new TransactionSettings(context, extraInfo);
					try {
						GenericPdu dataPdu = getDataPdu(context, contentLocation, transactionSettings);
						//If we have an MMS, we can parse it and send it upstream
						if (dataPdu instanceof MultimediaMessagePdu) {
							MultimediaMessagePdu mmsPdu = (MultimediaMessagePdu) dataPdu;
							//Undo binding so we don't send raw mms data upstream over mobile data
							connectivityManager.bindProcessToNetwork(null);
							sendUpstream(context, mmsPdu);
						}
					} catch (IOException e) {
						e.printStackTrace();
					}
				}
			});
		}
	}

	/**
	 * Removes the Line 1 number from a list of phone numbers.
	 * @param context The context in which the receiver is running.
	 * @param values The values to check for the line 1 number in.
	 * @return
	 */
	private EncodedStringValue[] removeLine1Number(Context context, EncodedStringValue[] values) {
		TelephonyManager telephonyManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
		//TODO: deal with possible lack of phone permissions.
		String line1Number = telephonyManager.getLine1Number();
		ArrayList<Integer> numberIndexes = new ArrayList<Integer>();
		//Find where the number is located in the list.
		for (int i = 0; i < values.length; i++)	 {
			if (values[i].getString().equals(line1Number)) {
				numberIndexes.add(i);
			}
		}
		if (numberIndexes.size() == 0) {
			return values;
		}
		//Create a new list with that value removed.
		EncodedStringValue[] correctedValues = new EncodedStringValue[values.length - 1];
		int indexOffset = 0;
		for (int i = 0; i < values.length; i++) {
			if (numberIndexes.contains(i)) {
				//We must offset the indexes we're assigning to within correctedValues.
				indexOffset++;
			} else {
				correctedValues[i - indexOffset] = values[i];
			}
		}

		return correctedValues;
	}

	/**
	 * Gets the necessary network to send upstream.
	 *
	 * Specifically, we need to send on cellular, with MMS capabilities
	 * @param connectivityManager The system's connectivity manager.
	 * @param callback A callback for once the proper network has been acquired by the system.
	 */
	private void getNetwork(ConnectivityManager connectivityManager, ConnectivityManager.NetworkCallback callback) {
		NetworkRequest networkRequest = (new NetworkRequest.Builder())
				.addCapability(NetworkCapabilities.NET_CAPABILITY_MMS)
				.addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
				.build();
		connectivityManager.requestNetwork(networkRequest, callback);
	}

	/**
	 * Get a PDU with the MMS' data.
	 * @param context The context in which the receiver is running.
	 * @param contentLocation The remote location that the carrier is hosting the MMS content within.
	 * @param transactionSettings The TransactionSettings object representing the network/APN settings necessary to retrieve the MMS.
	 * @return The PDU with the extracted data.
	 * @throws IOException Thrown when there is an error in retrieving the data from the specified remote.
	 */
	private GenericPdu getDataPdu(Context context, String contentLocation, TransactionSettings transactionSettings) throws IOException {
		byte[] rawPdu = HttpUtils.httpConnection(context, 0, contentLocation, null, HttpUtils.HTTP_GET_METHOD, transactionSettings.isProxySet(), transactionSettings.getProxyAddress(), transactionSettings.getProxyPort());
		PduParser dataParser = new PduParser(rawPdu);
		GenericPdu dataPdu = dataParser.parse();

		return dataPdu;
	}

	/**
	 * Gets the to, cc, and bcc fields from a PDU header.
	 * @param context The context in which the receiver is running.
	 * @param headers The headers of the PDU to extract the addresses from.
	 * @return A map of the various 'to' fields in the MMS.
	 */
	private Map<Integer, EncodedStringValue[]> getToFields(Context context, PduHeaders headers) {
		HashMap<Integer, EncodedStringValue[]> addresses = new HashMap<>();
		//Iterate through all the different types of to fields and add them to the addresses set
		for (int addressType : TO_ADDRESS_TYPES) {
			EncodedStringValue[] rawFieldValues = headers.getEncodedStringValues(addressType);
			if (addressType != PduHeaders.TO && rawFieldValues != null) {
				EncodedStringValue[] fieldValues = removeLine1Number(context, rawFieldValues);
				addresses.put(addressType, fieldValues);
			} else {
				addresses.put(addressType, rawFieldValues);
			}
		}

		return addresses;
	}

	/**
	 * Gets the recipients from an MMS message.
	 *
	 * This is just a merged list of the result from 'getToFields'.
	 * @param context The context in which the receiver is running.
	 * @param pdu The PDU to extract the recipients from.
	 * @return The recipients for the message.
	 */
	private List<String> getRecipients(Context context, MultimediaMessagePdu pdu) {
		PduHeaders headers = pdu.getPduHeaders();
		Map<Integer, EncodedStringValue[]> addresses = getToFields(context, headers);
		List<String> recipients = new ArrayList<>();
		for (int addressType : TO_ADDRESS_TYPES) {
			EncodedStringValue[] addressesOfType = addresses.get(addressType);
			if (addressesOfType != null) {
				for (EncodedStringValue rawAddress : addressesOfType) {
					String address = rawAddress.getString();
					recipients.add(address);
				}
			}
		}

		return recipients;
	}

	/**
	 * Make a list of the parts within a PDU.
	 * @param pdu The PDU to extract the parts from.
	 * @return A list of the parts within the PDU.
	 */
	private List<String> makePartsList(MultimediaMessagePdu pdu) {
		PduBody body = pdu.getBody();
		int numParts = body.getPartsNum();
		List<String> partsList = new ArrayList<>();

		//Iterate through all the parts. If the type is supported, we add it to the list of parts.
		for (int i = 0; i < numParts; i++) {
			PduPart part = body.getPart(i);
			String contentType = new String(part.getContentType());
			if (ContentType.isSupportedType(contentType) && !ContentType.isDrmType(contentType) && !contentType.equals(ContentType.APP_SMIL)) {
				byte[] partData = part.getData();
				String b64String = Base64.encodeToString(partData, Base64.DEFAULT);
				partsList.add(b64String);
			}
		}

		return partsList;
	}

	/**
	 * Send the MMS upstream to the sms-pusher server.
	 * @param context The context in which the receiver is running.
	 * @param pdu The pdu that contains the MMS' data.
	 */
	private void sendUpstream(Context context, final MultimediaMessagePdu pdu) {
		List<String> recipients = getRecipients(context, pdu);
		final List<String> partsList = makePartsList(pdu);
		final JSONArray recipientsJSONArray = new JSONArray(recipients);
		final PartSender partSender = new PartSender(context, partsList, (String blockId) -> {
			Map<String, String> mmsComponentsMap = new HashMap<>();
			mmsComponentsMap.put(RECIPIENTS_KEY, recipientsJSONArray.toString());
			mmsComponentsMap.put(BLOCK_ID_KEY, blockId);
			String fromNumber = pdu.getFrom().toString();
			long timestamp = pdu.getDate();
			MessageSender.Message upstreamMessage = new MessageSender.Message(fromNumber, null, timestamp, mmsComponentsMap);
			MessageSender.sendMessageUpstream(upstreamMessage);
		});
		partSender.send(context);
	}
}
