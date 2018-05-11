package com.ollien.smspusher;

import com.google.firebase.messaging.FirebaseMessaging;
import com.google.firebase.messaging.RemoteMessage;

import java.util.Map;
import java.util.UUID;

/**
 * Created by nick on 3/4/18.
 */

public class MessageSender {
	private final static String SENDER_ID = "363587568570";
	private final static String SENDER_SUFFIX = "@gcm.googleapis.com";

	public static class Message {
		private final String PHONE_NUMBER_KEY = "phone_number";
		private final String TEXT_KEY = "message";
		private final String TIMESTAMP_KEY = "timestamp";

		private Map<String, String> otherData;
		private String phoneNumber;
		private String text;
		private long timestamp;

		public Message(String phoneNumber, String text, long timestamp) {
			this(phoneNumber, text, timestamp, null);
		}

		public Message(String phoneNumber, String text, long timestamp, Map<String, String> otherData) {
			this.phoneNumber = phoneNumber;
			this.text = text;
			this.timestamp = timestamp;
			this.otherData = otherData;
		}
	}

	private MessageSender(){}

	private static String generateMessageId() {
		UUID uuid = UUID.randomUUID();
		return uuid.toString();
	}

	protected static void sendMessageUpstream(Message message) {
		FirebaseMessaging firebaseMessaging = FirebaseMessaging.getInstance();
		RemoteMessage.Builder payloadBuilder = new RemoteMessage.Builder(SENDER_ID + SENDER_SUFFIX)
			.setMessageId(generateMessageId())
			.addData(message.PHONE_NUMBER_KEY, message.phoneNumber)
			.addData(message.TEXT_KEY, message.text)
			.addData(message.TIMESTAMP_KEY, String.valueOf(message.timestamp))
			.setTtl(0); //Hotfix for issue with ACKs.
		if (message.otherData != null) {
			for (String key : message.otherData.keySet()) {
				String value = message.otherData.get(key);
				payloadBuilder.addData(key, value);
			}
		}
		RemoteMessage payload = payloadBuilder.build();
		firebaseMessaging.send(payload);
	}
}
