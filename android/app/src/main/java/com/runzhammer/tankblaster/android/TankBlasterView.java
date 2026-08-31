package com.runzhammer.tankblaster.android;

import android.app.AlertDialog;
import android.content.Context;
import android.util.Log;

import com.runzhammer.tankblaster.mobile.EbitenView;

public final class TankBlasterView extends EbitenView {
    public TankBlasterView(Context context) {
        super(context);
    }

    @Override
    protected void onErrorOnGameUpdate(Exception e) {
        Log.e("TankBlaster", "Game update failed", e);
        post(() -> new AlertDialog.Builder(getContext())
                .setTitle("Tank Blaster Startfehler")
                .setMessage(e.toString())
                .setPositiveButton(android.R.string.ok, null)
                .show());
    }
}
