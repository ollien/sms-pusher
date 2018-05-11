package com.ollien.smspusher;

import android.telephony.SmsManager;
import android.util.Log;

import com.google.firebase.messaging.FirebaseMessagingService;
import com.google.firebase.messaging.RemoteMessage;

/**
 * Handles receipt of FCM messages.
 */
public class FirebaseMessageService extends FirebaseMessagingService {
	/**
	 * Sends a SMS when a message is received from FCM.
	 * @param remoteMessage The message encoded to be sent.
	 */
	@Override
	public void onMessageReceived(RemoteMessage remoteMessage) {
		String phoneNumber = remoteMessage.getData().get("phone_number");
		String message = remoteMessage.getData().get("message");
		SmsManager smsManager = SmsManager.getDefault();
		smsManager.sendTextMessage(phoneNumber, null, message, null, null);
	}
}
