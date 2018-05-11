package com.ollien.smspusher;

import android.Manifest;
import android.content.SharedPreferences;
import android.support.v4.app.ActivityCompat;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;

import com.android.volley.Request;
import com.android.volley.RequestQueue;
import com.android.volley.Response;
import com.android.volley.VolleyError;
import com.android.volley.toolbox.StringRequest;
import com.android.volley.toolbox.Volley;

import org.json.JSONException;
import org.json.JSONObject;

import java.net.CookieHandler;
import java.net.CookieManager;
import java.net.HttpCookie;
import java.net.MalformedURLException;
import java.net.URISyntaxException;
import java.net.URL;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * The main activity of the application.
 */
public class MainActivity extends AppCompatActivity {

	protected static final String PREFS_KEY = "SMSPusherPrefs";
	protected static final String SESSION_ID_PREFS_KEY = "session_id";
	protected static final String DEVICE_ID_PREFS_KEY = "device_id";
	protected static final String HOST_URL_PREFS_KEY = "host_url";
	protected static final String FCM_TOKEN_PREFS_KEY = "fcm_token";

	private RequestQueue queue;
	private EditText hostField;
	private EditText usernameField;
	private EditText passwordField;
	private TextView deviceIDDisplay;
	private SharedPreferences prefs;
	private SharedPreferences.Editor prefsEditor;
	private CookieManager cookieManager;

	/**
	 * Handles the creation of the activity.
	 * @param savedInstanceState The saved state of the activity..
	 */
	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);
		ActivityCompat.requestPermissions(this, new String[]{Manifest.permission.RECEIVE_SMS}, 0);
		queue = Volley.newRequestQueue(this);
		cookieManager = new CookieManager();
		CookieHandler.setDefault(cookieManager);
		hostField = (EditText)findViewById(R.id.register_host);
		usernameField = (EditText)findViewById(R.id.register_username);
		passwordField = (EditText)findViewById(R.id.register_password);
		deviceIDDisplay = (TextView)findViewById(R.id.device_id_text_view);
		prefs = getSharedPreferences(PREFS_KEY, MODE_PRIVATE);
		prefsEditor = prefs.edit();
		updateDeviceIDDisplay();
	}

	/**
	 * Registers the user based on the entered information.
	 * @param v The view of the button that was pressed.
	 */
	protected void handleRegisterButtonPress(View v) {
		Toast.makeText(this, R.string.registering_toast, Toast.LENGTH_SHORT).show();
		try {
			final URL host = new URL(hostField.getText().toString());
			Response.Listener<String> resListener = new Response.Listener<String>() {
				@Override
				public void onResponse(String response) {
					try {
						registerDevice(host, new Response.Listener<String>() {
							@Override
							public void onResponse(String response) {
								FirebaseIdService.updateTokenOnServer(prefs, queue, new Response.Listener<String>() {
									@Override
									public void onResponse(String response) {
										Toast.makeText(MainActivity.this, R.string.registered_toast, Toast.LENGTH_SHORT).show();
										updateDeviceIDDisplay();
									}
								}, new Response.ErrorListener() {
									@Override
									public void onErrorResponse(VolleyError e) {
										Toast.makeText(MainActivity.this, R.string.connection_error_toast, Toast.LENGTH_SHORT).show();
										Log.e("SMSPusher", e.toString());
									}
								});
							}
						}, new Response.ErrorListener() {
							@Override
							public void onErrorResponse(VolleyError e) {
								Toast.makeText(MainActivity.this, R.string.connection_error_toast, Toast.LENGTH_SHORT).show();
								Log.e("SMSPusher", e.toString());
							}
						});
					}
					catch (MalformedURLException e) {
						Toast.makeText(MainActivity.this, R.string.invalid_host_toast, Toast.LENGTH_SHORT).show();
						Log.e("SMSPusher", e.toString());
					}
				}
			};
			if (prefs.getString(SESSION_ID_PREFS_KEY, "").equals("")) {
				String username = usernameField.getText().toString();
				String password = passwordField.getText().toString();
				authenticate(host, username, password, resListener, new Response.ErrorListener() {
					@Override
					public void onErrorResponse(VolleyError e) {
						Toast.makeText(MainActivity.this, R.string.connection_error_toast, Toast.LENGTH_SHORT).show();
						Log.e("SMSPusher", e.toString());
					}
				});
			}
			else {
				resListener.onResponse(null);
			}
		}
		catch (MalformedURLException e) {
			Toast.makeText(this, R.string.invalid_host_toast, Toast.LENGTH_SHORT).show();
			Log.e("SMSPusher", e.toString());
		}
	}

	/**
	 * Registers a device in the remote database.
	 *
	 * Assumes user is already authenticated. Authentication should be checked before use.
	 * @param host The host to register the device on.
	 * @param resListener A response listener for the server's response.
	 * @param errorListener An error listener for the server's response.
	 * @throws MalformedURLException Thrown when an invalid host is passed.
	 */
	private void registerDevice(final URL host, final Response.Listener<String> resListener, final Response.ErrorListener errorListener) throws MalformedURLException {
		final URL registerUrl = new URL(host, "/register_device");
		final HashMap<String, String> authMap = new HashMap<String, String>();
		authMap.put("session_id", prefs.getString(SESSION_ID_PREFS_KEY, ""));
		StringRequest req = new StringRequest(Request.Method.POST, registerUrl.toString(), new Response.Listener<String>() {
			@Override
			public void onResponse(String response) {
				JSONObject resJSON = null;
				try {
					resJSON = new JSONObject(response);
					String deviceID = resJSON.getString("device_id");
					prefsEditor.putString(DEVICE_ID_PREFS_KEY, deviceID);
					prefsEditor.putString(HOST_URL_PREFS_KEY, host.toString());
					prefsEditor.apply();
					if (resListener != null) {
						resListener.onResponse(deviceID);
					}
				}
				catch (JSONException e) {
					Log.e("SMSPusher", e.toString());
					if (errorListener != null) {
						errorListener.onErrorResponse(new VolleyError(e));
					}
				}
			}
		}, errorListener) {
			protected Map<String, String> getParams() {
				return authMap;
			}
		};
		queue.add(req);
	}

	/**
	 * Authenticates the user against the server.
	 * @param host The host to verify the registration against.
	 * @param username The username of the user.
	 * @param password The password of the user.
	 * @param resListener A response listener for the server's response.
	 * @param errorListener An error listener for the server's response.
	 * @throws MalformedURLException Thrown when an invalid host is passed.
	 */
	private void authenticate(URL host, String username, String password, final Response.Listener<String> resListener, final Response.ErrorListener errorListener) throws MalformedURLException {
		final URL authURL = new URL(host, "/authenticate");
		final HashMap<String, String> authMap = new HashMap<String, String>();
		authMap.put("username", username);
		authMap.put("password", password);
		StringRequest req = new StringRequest(Request.Method.POST, authURL.toString(), new Response.Listener<String>() {
			@Override
			public void onResponse(String response) {
				try {
					List<HttpCookie> cookies = cookieManager.getCookieStore().get(authURL.toURI());
					String sessionID = null;
					for (HttpCookie cookie : cookies) {
						if (cookie.getName().equals("session")) {
							sessionID = cookie.getValue();
						}
					}
					if (sessionID == null) {
						throw new NoSuchFieldException("No session key exists.");
					}

					prefsEditor.putString(SESSION_ID_PREFS_KEY, sessionID);
					prefsEditor.apply();
					if (resListener != null) {
						resListener.onResponse(sessionID);
					}
				}
				catch (NoSuchFieldException | URISyntaxException e) {
					Log.e("SMSPusher", e.toString());
					if (errorListener != null) {
						errorListener.onErrorResponse(new VolleyError(e));
					}
				}
			}
		}, errorListener) {
			protected Map<String, String> getParams() {
				return authMap;
			}
		};
		queue.add(req);
	}

	/**
	 * Updates the displayed device ID.
	 */
	private void updateDeviceIDDisplay(){
		String deviceID = prefs.getString(DEVICE_ID_PREFS_KEY, "");
		if (deviceID.equals("")) {
			deviceIDDisplay.setText(R.string.no_device_id);
		}
		else {
			deviceIDDisplay.setText("Device ID: " + deviceID);
		}
	}
}
