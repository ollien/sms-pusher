package com.ollien.smspusher;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.widget.Toast;

import java.util.UUID;

public class SMSReceiver extends BroadcastReceiver {

	@Override
	public void onReceive(Context context, Intent intent) {
		Toast.makeText(context, "Received!", Toast.LENGTH_LONG).show();
	}

	private String generateMessageId() {
		UUID uuid = UUID.randomUUID();
		return uuid.toString();
	}
}
