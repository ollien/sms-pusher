package com.ollien.smspusher;

import com.google.firebase.messaging.FirebaseMessaging;
import com.google.firebase.messaging.RemoteMessage;

import java.util.Map;
import java.util.UUID;

/**
 * Sends messages upstream via FCM.
 */
public class MessageSender {
	private final static String SENDER_ID = "363587568570";
	private final static String SENDER_SUFFIX = "@gcm.googleapis.com";

	/**
	 * A single message to be sent upstream.
	 */
	public static class Message {
		private final String PHONE_NUMBER_KEY = "phone_number";
		private final String TEXT_KEY = "message";
		private final String TIMESTAMP_KEY = "timestamp";

		private Map<String, String> otherData;
		private String phoneNumber;
		private String text;
		private long timestamp;

		/**
		 * Constructs a message with the basic fields - primarily intended for SMS message.
		 * @param phoneNumber The phone number of the message being sent.
		 * @param text The text of the message being sent.
		 * @param timestamp The timestamp of the message being sent.
		 */
		public Message(String phoneNumber, String text, long timestamp) {
			this(phoneNumber, text, timestamp, null);
		}

		/**
		 * Constructs a message with the basic fields
		 * @param phoneNumber The phone number of the message being sent.
		 * @param text The text of the message being sent.
		 * @param timestamp The timestamp of the message being sent.
		 * @param otherData Any other data that should be sent upstream, such as MMS info.
		 */
		public Message(String phoneNumber, String text, long timestamp, Map<String, String> otherData) {
			this.phoneNumber = phoneNumber;
			this.text = text;
			this.timestamp = timestamp;
			this.otherData = otherData;
		}
	}

	/**
	 * Private constructor - makes MessageSender uninitializeable.
	 */
	private MessageSender(){}

	/**
	 * Generate an ID for a message
	 * @return A new message ID.
	 */
	private static String generateMessageId() {
		UUID uuid = UUID.randomUUID();
		return uuid.toString();
	}

	/**
	 * Send a message upstream to the server via FCM.
	 * @param message The message to be sent upstream.
	 */
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
