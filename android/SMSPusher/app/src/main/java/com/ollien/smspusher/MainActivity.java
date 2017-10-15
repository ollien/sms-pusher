package com.ollien.smspusher;

import android.Manifest;
import android.content.SharedPreferences;
import android.support.v4.app.ActivityCompat;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.widget.EditText;
import android.widget.Toast;

import com.android.volley.RequestQueue;
import com.android.volley.toolbox.Volley;

public class MainActivity extends AppCompatActivity {

	private final String PREFS_KEY = "SMSPusherPrefs";

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
}
