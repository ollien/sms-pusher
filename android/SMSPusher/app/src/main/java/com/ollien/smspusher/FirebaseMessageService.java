package com.ollien.smspusher;

import android.util.Log;

import com.google.firebase.messaging.FirebaseMessagingService;
import com.google.firebase.messaging.RemoteMessage;


public class FirebaseMessageService extends FirebaseMessagingService {
	@Override
	public void onMessageReceived(RemoteMessage message) {
		for (String key : message.getData().keySet()) {
			Log.d("SMSPusherRemoteMessage", message.getData().get(key));
		}
	}
}
