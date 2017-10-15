package com.ollien.smspusher;

import android.Manifest;
import android.content.SharedPreferences;
import android.support.v4.app.ActivityCompat;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.widget.EditText;
import android.widget.Toast;

import com.android.volley.Request;
import com.android.volley.RequestQueue;
import com.android.volley.Response;
import com.android.volley.VolleyError;
import com.android.volley.toolbox.StringRequest;
import com.android.volley.toolbox.Volley;

import org.json.JSONException;
import org.json.JSONObject;

import java.net.MalformedURLException;
import java.net.URL;
import java.util.HashMap;
import java.util.Map;

public class MainActivity extends AppCompatActivity {

	protected final String PREFS_KEY = "SMSPusherPrefs";
	protected final String SESSION_ID_PREFS_KEY = "session_id";

    private RequestQueue queue;
    private EditText hostField;
	private EditText usernameField;
	private EditText passwordField;
    private SharedPreferences prefs;
	private SharedPreferences.Editor prefsEditor;

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);
		ActivityCompat.requestPermissions(this, new String[]{Manifest.permission.RECEIVE_SMS}, 0);
		queue = Volley.newRequestQueue(this);
		hostField = (EditText)findViewById(R.id.register_host);
		usernameField = (EditText)findViewById(R.id.register_username);
		passwordField = (EditText)findViewById(R.id.register_password);
		prefs = getSharedPreferences(PREFS_KEY, MODE_PRIVATE);
        prefsEditor = prefs.edit();
	}

	private void authenticate(URL host, String username, String password, final Response.Listener<String> resListener, final Response.ErrorListener errorListener) throws MalformedURLException {
		URL authURL = new URL(host, "/authenticate");
		final HashMap<String, String> authMap = new HashMap<String, String>();
        authMap.put("username", username);
		authMap.put("password", password);
		StringRequest req = new StringRequest(Request.Method.POST, authURL.toString(), new Response.Listener<String>() {
			@Override
			public void onResponse(String response) {
				try {
					JSONObject resJSON = new JSONObject(response);
					String sessionID = resJSON.getString("session_id");
					prefsEditor.putString(SESSION_ID_PREFS_KEY, sessionID);
					prefsEditor.apply();
                    if (resListener != null) {
						resListener.onResponse(sessionID);
					}
				} catch (JSONException e) {
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
}
