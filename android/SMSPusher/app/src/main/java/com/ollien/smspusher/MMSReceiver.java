package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.database.Cursor;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkInfo;
import android.net.NetworkRequest;
import android.net.Uri;
import android.os.Bundle;
import android.os.TransactionTooLargeException;
import android.provider.Telephony;
import android.telephony.CarrierConfigManager;
import android.util.Log;

import com.android.mms.transaction.HttpUtils;
import com.android.mms.transaction.TransactionSettings;
import com.google.android.mms_clone.pdu.GenericPdu;
import com.google.android.mms_clone.pdu.MultimediaMessagePdu;
import com.google.android.mms_clone.pdu.NotificationInd;
import com.google.android.mms_clone.pdu.PduBody;
import com.google.android.mms_clone.pdu.PduHeaders;
import com.google.android.mms_clone.pdu.PduParser;
import com.google.android.mms_clone.pdu.PduPart;
import com.google.android.mms_clone.pdu.PduPersister;

import java.io.IOException;

/**
 * Created by nick on 12/19/17.
 */

public class MMSReceiver extends BroadcastReceiver {
	@Override
	public void onReceive(final Context context, Intent intent) {
		Bundle extras = intent.getExtras();
		byte[] data = extras.getByteArray("data");
		PduParser parser = new PduParser(data);
		GenericPdu genericPdu = parser.parse();
		if (genericPdu.getMessageType() == PduHeaders.MESSAGE_TYPE_NOTIFICATION_IND) {
			NotificationInd pdu = (NotificationInd) genericPdu;
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
						byte[] rawPdu = HttpUtils.httpConnection(context, 0, contentLocation, null, HttpUtils.HTTP_GET_METHOD, transactionSettings.isProxySet(), transactionSettings.getProxyAddress(), transactionSettings.getProxyPort());
						PduParser dataParser = new PduParser(rawPdu);
						GenericPdu dataPdu = dataParser.parse();
						if (dataPdu instanceof MultimediaMessagePdu) {
							MultimediaMessagePdu mmsPdu = (MultimediaMessagePdu) dataPdu;
							PduBody body = mmsPdu.getBody();
							String bytes = "";
							for (int i = 0; i < body.getPartsNum(); i++) {
								PduPart part = body.getPart(i);
								byte[] partData = part.getData();
								for (Byte b : partData) {
									bytes += b.toString() + ",";
								}
							}
							Log.d("SMSPusher", "done");
						}

					} catch (IOException e) {
						e.printStackTrace();
					}
				}
			});
//			TransactionSettings transactionSettings = new TransactionSettings(context, )
		}
	}

	private void getNetwork(ConnectivityManager connectivityManager, ConnectivityManager.NetworkCallback callback) {
		NetworkRequest networkRequest = (new NetworkRequest.Builder()).addCapability(NetworkCapabilities.NET_CAPABILITY_MMS).build();
		connectivityManager.requestNetwork(networkRequest, callback);
	}
}
