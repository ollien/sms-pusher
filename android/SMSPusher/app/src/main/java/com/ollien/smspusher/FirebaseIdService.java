package com.ollien.smspusher;

import android.content.SharedPreferences;
import android.util.Log;

import com.android.volley.Request;
import com.android.volley.RequestQueue;
import com.android.volley.Response;
import com.android.volley.VolleyError;
import com.android.volley.toolbox.StringRequest;
import com.android.volley.toolbox.Volley;
import com.google.firebase.iid.FirebaseInstanceId;
import com.google.firebase.iid.FirebaseInstanceIdService;

import java.net.MalformedURLException;
import java.net.URL;
import java.util.HashMap;
import java.util.Map;
import java.util.NoSuchElementException;

/**
 * Id service to handle FCM registration.
 */
public class FirebaseIdService extends FirebaseInstanceIdService {

	RequestQueue queue;
	SharedPreferences prefs;
	SharedPreferences.Editor prefsEditor;

	/**
	 * Initializes the ID service.
	 */
	@Override
	public void onCreate() {
		queue = Volley.newRequestQueue(this);
		prefs = getSharedPreferences(MainActivity.PREFS_KEY, MODE_PRIVATE);
		prefsEditor = prefs.edit();
	}

	/**
	 * Sends a new token upstream when a refresh is required.
	 */
	@Override
	public void onTokenRefresh() {
		String token = FirebaseInstanceId.getInstance().getToken();
		prefsEditor.putString(MainActivity.FCM_TOKEN_PREFS_KEY, token);
		String hostURL = prefs.getString(MainActivity.HOST_URL_PREFS_KEY, "");
		String deviceID = prefs.getString(MainActivity.DEVICE_ID_PREFS_KEY, "");
		if (!hostURL.equals("") && !deviceID.equals("")) {
			updateTokenOnServer();
		}
	}

	/**
	 * Updates the FCM token on the server with a default error handler.
	 */
	protected void updateTokenOnServer() {
		updateTokenOnServer(prefs, queue, null, new Response.ErrorListener() {
			@Override
			public void onErrorResponse(VolleyError e) {
				Log.e("SMSPusher", e.toString());
			}
		});
	}

	/**
	 * Updates the FCM token on the server.
	 * @param prefs The app's SharedPreferences
	 * @param queue The request queue to process the request within.
	 * @param resListener A response listener for the server's response.
	 * @param errorListener An error listener for the server's response.
	 */
	protected static void updateTokenOnServer(SharedPreferences prefs, RequestQueue queue, Response.Listener<String> resListener, Response.ErrorListener errorListener) {
		String hostURL = prefs.getString(MainActivity.HOST_URL_PREFS_KEY, "");
		String sessionID = prefs.getString(MainActivity.SESSION_ID_PREFS_KEY, "");
		String deviceID = prefs.getString(MainActivity.DEVICE_ID_PREFS_KEY, "");
		String token = prefs.getString(MainActivity.FCM_TOKEN_PREFS_KEY, "");
		if (hostURL.equals("")) {
			throw new NoSuchElementException("hostURL is not defined within preferences");
		}
		else if (deviceID.equals("")) {
			throw new NoSuchElementException("deviceID is not defined within preferences");
		}
		else if (token.equals("")) {
			token = FirebaseInstanceId.getInstance().getToken();
		}
		URL updateURL;
		try {
			updateURL = new URL(new URL(hostURL), "/set_fcm_id");
			final HashMap<String, String> reqMap = new HashMap<>();
			reqMap.put("fcm_id", token);
			reqMap.put("device_id", deviceID);
			reqMap.put("session_id", sessionID);
			StringRequest req = new StringRequest(Request.Method.POST, updateURL.toString(),  resListener, errorListener)
			{
				protected Map<String, String> getParams() {
					return reqMap;
				}
			};
			queue.add(req);
		} catch (MalformedURLException e) {
			errorListener.onErrorResponse(new VolleyError(e));
		}
	}

}
