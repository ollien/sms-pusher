package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.telephony.SmsMessage;
import android.widget.Toast;

import com.google.firebase.messaging.FirebaseMessaging;
import com.google.firebase.messaging.RemoteMessage;

import java.util.UUID;

public class SMSReceiver extends BroadcastReceiver {

	private final String SENDER_ID = "363587568570";

	@Override
	public void onReceive(Context context, Intent intent) {
		Toast.makeText(context, "Received!", Toast.LENGTH_LONG).show();
	}

	private String generateMessageId() {
		UUID uuid = UUID.randomUUID();
		return uuid.toString();
	}

	private void sendMessageUpstream(SmsMessage message) {
		FirebaseMessaging firebaseMessaging = FirebaseMessaging.getInstance();
		RemoteMessage payload = new RemoteMessage.Builder(SENDER_ID + "@gc.googleapis.com")
				.setMessageId(generateMessageId())
				.addData("phone_number", message.getOriginatingAddress())
				.addData("message", message.getMessageBody())
				.addData("timestamp", String.valueOf(message.getTimestampMillis()))
				.build();
		firebaseMessaging.send(payload);
	}
}
