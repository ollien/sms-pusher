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

import com.android.mms.transaction.HttpUtils;
import com.android.mms.transaction.TransactionSettings;
import com.google.android.mms_clone.pdu.GenericPdu;
import com.google.android.mms_clone.pdu.MultimediaMessagePdu;
import com.google.android.mms_clone.pdu.NotificationInd;
import com.google.android.mms_clone.pdu.PduBody;
import com.google.android.mms_clone.pdu.PduHeaders;
import com.google.android.mms_clone.pdu.PduParser;

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
						GenericPdu dataPdu = getDataPdu(context, contentLocation, transactionSettings);
						if (dataPdu instanceof MultimediaMessagePdu) {
							MultimediaMessagePdu mmsPdu = (MultimediaMessagePdu) dataPdu;
							PduBody body = mmsPdu.getBody();
						}
					} catch (IOException e) {
						e.printStackTrace();
					}
				}
			});
		}
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
}
