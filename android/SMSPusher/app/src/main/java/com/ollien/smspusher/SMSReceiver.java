package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.telephony.SmsMessage;

import com.google.firebase.messaging.FirebaseMessaging;
import com.google.firebase.messaging.RemoteMessage;

import java.util.UUID;

public class SMSReceiver extends BroadcastReceiver {

	private final String SENDER_ID = "363587568570";

	@Override
	public void onReceive(Context context, Intent intent) {
		byte[][] pdus = (byte[][])intent.getSerializableExtra("pdus");
		String format = intent.getStringExtra("format");
		for (byte[] pdu : pdus) {
			SmsMessage message = SmsMessage.createFromPdu(pdu, format);
			sendMessageUpstream(message);
		}

	}

	private String generateMessageId() {
		UUID uuid = UUID.randomUUID();
		return uuid.toString();
	}

	private void sendMessageUpstream(SmsMessage message) {
		FirebaseMessaging firebaseMessaging = FirebaseMessaging.getInstance();
		RemoteMessage payload = new RemoteMessage.Builder(SENDER_ID + "@gcm.googleapis.com")
				.setMessageId(generateMessageId())
				.addData("phone_number", message.getOriginatingAddress())
				.addData("message", message.getMessageBody())
				.addData("timestamp", String.valueOf(message.getTimestampMillis()))
				.build();
		firebaseMessaging.send(payload);
	}
}
