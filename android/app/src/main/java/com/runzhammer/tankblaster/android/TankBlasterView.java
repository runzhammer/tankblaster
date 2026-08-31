package com.runzhammer.tankblaster.android;

import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.util.Log;
import android.view.inputmethod.BaseInputConnection;
import android.view.inputmethod.EditorInfo;
import android.view.inputmethod.InputConnection;
import android.view.inputmethod.InputMethodManager;

import com.runzhammer.tankblaster.mobile.EbitenView;
import com.runzhammer.tankblaster.mobile.Mobile;

public final class TankBlasterView extends EbitenView {
    private final Runnable keyboardSync = new Runnable() {
        @Override
        public void run() {
            syncSoftKeyboard();
            syncClipboard();
            postDelayed(this, 120);
        }
    };
    private boolean keyboardVisible;

    public TankBlasterView(Context context) {
        super(context);
        setFocusable(true);
        setFocusableInTouchMode(true);
        post(keyboardSync);
    }

    @Override
    public boolean onCheckIsTextEditor() {
        return true;
    }

    @Override
    public InputConnection onCreateInputConnection(EditorInfo outAttrs) {
        outAttrs.inputType = android.text.InputType.TYPE_CLASS_TEXT
                | android.text.InputType.TYPE_TEXT_FLAG_CAP_SENTENCES
                | android.text.InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS;
        outAttrs.imeOptions = EditorInfo.IME_ACTION_DONE
                | EditorInfo.IME_FLAG_NO_EXTRACT_UI
                | EditorInfo.IME_FLAG_NO_FULLSCREEN;
        return new GameInputConnection(this);
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

    private void syncSoftKeyboard() {
        boolean shouldShow = Mobile.playerNameInputActive();
        if (shouldShow == keyboardVisible) {
            return;
        }
        keyboardVisible = shouldShow;
        InputMethodManager imm = (InputMethodManager)getContext().getSystemService(Context.INPUT_METHOD_SERVICE);
        if (imm == null) {
            return;
        }
        if (shouldShow) {
            requestFocus();
            postDelayed(() -> imm.showSoftInput(this, InputMethodManager.SHOW_IMPLICIT), 50);
            return;
        }
        imm.hideSoftInputFromWindow(getWindowToken(), 0);
        clearFocus();
    }

    private void syncClipboard() {
        String text = Mobile.pendingClipboardText();
        if (text == null || text.isEmpty()) {
            return;
        }
        ClipboardManager clipboard = (ClipboardManager)getContext().getSystemService(Context.CLIPBOARD_SERVICE);
        if (clipboard != null) {
            clipboard.setPrimaryClip(ClipData.newPlainText("Tank Blaster", text));
        }
        Mobile.clearPendingClipboardText(text);
    }

    private static final class GameInputConnection extends BaseInputConnection {
        private final TankBlasterView view;
        private final StringBuilder text = new StringBuilder();
        private int composingStart = -1;
        private int composingEnd = -1;

        GameInputConnection(TankBlasterView view) {
            super(view, true);
            this.view = view;
        }

        @Override
        public boolean commitText(CharSequence text, int newCursorPosition) {
            if (text == null) {
                return true;
            }
            if ("\n".contentEquals(text)) {
                finishInput();
                return true;
            }
            replaceComposing(text.toString());
            clearComposing();
            publishText();
            return true;
        }

        @Override
        public boolean setComposingText(CharSequence text, int newCursorPosition) {
            if (text == null) {
                text = "";
            }
            replaceComposing(text.toString());
            publishText();
            return true;
        }

        @Override
        public boolean finishComposingText() {
            clearComposing();
            publishText();
            return true;
        }

        @Override
        public boolean deleteSurroundingText(int beforeLength, int afterLength) {
            if (hasComposing()) {
                text.delete(composingStart, composingEnd);
                clearComposing();
            } else if (text.length() > 0) {
                int last = text.offsetByCodePoints(text.length(), -1);
                text.delete(last, text.length());
            }
            publishText();
            return true;
        }

        @Override
        public boolean deleteSurroundingTextInCodePoints(int beforeLength, int afterLength) {
            return deleteSurroundingText(beforeLength, afterLength);
        }

        @Override
        public boolean performEditorAction(int editorAction) {
            if (editorAction == EditorInfo.IME_ACTION_DONE
                    || editorAction == EditorInfo.IME_ACTION_GO
                    || editorAction == EditorInfo.IME_ACTION_SEND
                    || editorAction == EditorInfo.IME_ACTION_UNSPECIFIED) {
                finishInput();
                return true;
            }
            return super.performEditorAction(editorAction);
        }

        @Override
        public boolean sendKeyEvent(android.view.KeyEvent event) {
            if (event.getKeyCode() == android.view.KeyEvent.KEYCODE_ENTER && event.getAction() == android.view.KeyEvent.ACTION_UP) {
                finishInput();
                return true;
            }
            return view.dispatchKeyEvent(event);
        }

        private void finishInput() {
            clearComposing();
            publishText();
            Mobile.finishPlayerNameInput();
            InputMethodManager imm = (InputMethodManager)view.getContext().getSystemService(Context.INPUT_METHOD_SERVICE);
            if (imm != null) {
                imm.hideSoftInputFromWindow(view.getWindowToken(), 0);
            }
            view.keyboardVisible = false;
        }

        private void replaceComposing(String value) {
            if (hasComposing()) {
                text.replace(composingStart, composingEnd, value);
                composingEnd = composingStart + value.length();
                return;
            }
            composingStart = text.length();
            text.append(value);
            composingEnd = text.length();
        }

        private boolean hasComposing() {
            return composingStart >= 0 && composingEnd >= composingStart && composingEnd <= text.length();
        }

        private void clearComposing() {
            composingStart = -1;
            composingEnd = -1;
        }

        private void publishText() {
            Mobile.setPlayerNameText(text.toString());
        }
    }
}
