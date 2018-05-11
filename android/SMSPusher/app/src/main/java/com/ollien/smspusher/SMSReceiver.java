package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.telephony.SmsMessage;

public class SMSReceiver extends BroadcastReceiver {


	@Override
	public void onReceive(Context context, Intent intent) {
		byte[][] pdus = (byte[][])intent.getSerializableExtra("pdus");
		String format = intent.getStringExtra("format");
		for (byte[] pdu : pdus) {
			SmsMessage message = SmsMessage.createFromPdu(pdu, format);
			sendMessageUpstream(message);
		}

	}

	private void sendMessageUpstream(SmsMessage message) {
		String phoneNumber = message.getOriginatingAddress();
		String messageText = message.getMessageBody();
		long timestamp = message.getTimestampMillis()/1000;
		MessageSender.Message upstreamMessage = new MessageSender.Message(phoneNumber, messageText, timestamp);
		MessageSender.sendMessageUpstream(upstreamMessage);
	}
}
