package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkInfo;
import android.net.NetworkRequest;
import android.os.Bundle;
import android.telephony.TelephonyManager;
import android.util.Base64;

import com.android.mms.transaction.HttpUtils;
import com.android.mms.transaction.TransactionSettings;
import com.google.android.mms_clone.ContentType;
import com.google.android.mms_clone.pdu.EncodedStringValue;
import com.google.android.mms_clone.pdu.GenericPdu;
import com.google.android.mms_clone.pdu.MultimediaMessagePdu;
import com.google.android.mms_clone.pdu.NotificationInd;
import com.google.android.mms_clone.pdu.PduBody;
import com.google.android.mms_clone.pdu.PduHeaders;
import com.google.android.mms_clone.pdu.PduParser;
import com.google.android.mms_clone.pdu.PduPart;

import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Created by nick on 12/19/17.
 */

public class MMSReceiver extends BroadcastReceiver {
	//Holds the different types of headers that represent the "to" field.
	private static final int[] TO_ADDRESS_TYPES = {PduHeaders.TO, PduHeaders.BCC, PduHeaders.CC};
	private static final String CONTENT_TYPE_KEY = "type";
	private static final String DATA_KEY = "data";

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
					NetworkInfo info = connectivityManager.getNetworkInfo(network);
					String extraInfo = info.getExtraInfo();
					TransactionSettings transactionSettings = new TransactionSettings(context, extraInfo);
					try {
						GenericPdu dataPdu = getDataPdu(context, contentLocation, transactionSettings);
						//If we have an MMS, we can parse it and send it upstream
						if (dataPdu instanceof MultimediaMessagePdu) {
							MultimediaMessagePdu mmsPdu = (MultimediaMessagePdu) dataPdu;
							PduBody body = mmsPdu.getBody();
							//TODO: Send upstream
						}
					} catch (IOException e) {
						e.printStackTrace();
					}
				}
			});
		}
	}

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

	private void getNetwork(ConnectivityManager connectivityManager, ConnectivityManager.NetworkCallback callback) {
		NetworkRequest networkRequest = (new NetworkRequest.Builder()).addCapability(NetworkCapabilities.NET_CAPABILITY_MMS).build();
		connectivityManager.requestNetwork(networkRequest, callback);
	}

	private GenericPdu getDataPdu(Context context, String contentLocation, TransactionSettings transactionSettings) throws IOException {
		byte[] rawPdu = HttpUtils.httpConnection(context, 0, contentLocation, null, HttpUtils.HTTP_GET_METHOD, transactionSettings.isProxySet(), transactionSettings.getProxyAddress(), transactionSettings.getProxyPort());
		PduParser dataParser = new PduParser(rawPdu);
		GenericPdu dataPdu = dataParser.parse();

		return dataPdu;
	}

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

	private List<String> getRecipients(Context context, MultimediaMessagePdu pdu) {
		PduHeaders headers = pdu.getPduHeaders();
		Map<Integer, EncodedStringValue[]> addresses = getToFields(context, headers);
		List<String> recipients = new ArrayList<>();
		for (int addressType : TO_ADDRESS_TYPES) {
			EncodedStringValue[] addressesOfType = addresses.get(addressType);
			for (EncodedStringValue rawAddress : addressesOfType) {
				String address = rawAddress.getString();
				recipients.add(address);
			}
		}

		return recipients;
	}

	private List<String> makePartsList(MultimediaMessagePdu pdu) {
		PduBody body = pdu.getBody();
		int numParts = body.getPartsNum();
		List<String> partsList = new ArrayList<>();

		//Iterate through all the parts. If the type is supported, we add it to the list of parts.
		for (int i = 0; i < numParts; i++) {
			PduPart part = body.getPart(i);
			String contentType = new String(part.getContentType());
			if (ContentType.isSupportedType(contentType) && !ContentType.isDrmType(contentType)) {
				try {
					byte[] partData = part.getData();
					String b64String = Base64.encodeToString(partData, Base64.DEFAULT);
					JSONObject partJSON = new JSONObject();
					partJSON.put(CONTENT_TYPE_KEY, contentType);
					partJSON.put(DATA_KEY, b64String);
					partsList.add(partJSON.toString());
				} catch (JSONException e) {
					e.printStackTrace();
					return null;
				}
			}
		}

		return partsList;
	}
}
