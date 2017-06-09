package com.ollien.smspusher;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;

import com.google.firebase.iid.FirebaseInstanceId;
import com.google.firebase.iid.FirebaseInstanceIdService;

public class FirebaseIdService extends FirebaseInstanceIdService {
	public FirebaseIdService() {
	}

	@Override
	public void onTokenRefresh() {
		//TODO: Send this to server
		//For now, just leaving it in here for reference.
		String token = FirebaseInstanceId.getInstance().getToken();
	}

}
