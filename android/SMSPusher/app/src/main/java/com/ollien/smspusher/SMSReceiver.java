package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.telephony.SmsMessage;

/**
 * Receives SMS messages from the system to send upstream.
 */
public class SMSReceiver extends BroadcastReceiver {

	/**
	 * Sends an SMS message upstream on receipt.
	 * @param context The context in which the receiver is running.
	 * @param intent The intent being processed
	 */
	@Override
	public void onReceive(Context context, Intent intent) {
		byte[][] pdus = (byte[][])intent.getSerializableExtra("pdus");
		String format = intent.getStringExtra("format");
		for (byte[] pdu : pdus) {
			SmsMessage message = SmsMessage.createFromPdu(pdu, format);
			sendMessageUpstream(message);
		}

	}

	/**
	 * Send the SMS upstream via FCM.
	 * @param message The SMS to be sent.
	 */
	private void sendMessageUpstream(SmsMessage message) {
		String phoneNumber = message.getOriginatingAddress();
		String messageText = message.getMessageBody();
		long timestamp = message.getTimestampMillis()/1000;
		MessageSender.Message upstreamMessage = new MessageSender.Message(phoneNumber, messageText, timestamp);
		MessageSender.sendMessageUpstream(upstreamMessage);
	}
}
