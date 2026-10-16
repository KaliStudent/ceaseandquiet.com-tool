# Local beta checklist

Use a new Chrome development profile and synthetic details. Use the package manifest's source commit for final review.

## Install and open

- [ ] Extract the package. Start preview/cq-preview.exe and open http://127.0.0.1:8787/ in Chrome.
- [ ] In chrome://extensions, enable Developer mode, select Load unpacked, and choose extension/.
- [ ] Confirm the extension loads without errors. Its permissions are storage, activeTab, and scripting, with no permanent host permissions.
- [ ] Open the extension from its toolbar action while the local preview tab is active.

## Vault lifecycle

- [ ] Enter Taylor Example, taylor@example.test, and a sample region. Set a unique passphrase with at least 12 characters.
- [ ] Encrypt and save. The revealed fields clear and the popup locks.
- [ ] Close and reopen the popup. The vault remains locked and no profile values are visible.
- [ ] Try a wrong passphrase. The popup remains locked and clears the passphrase.
- [ ] Unlock correctly. Check the revealed sample values.
- [ ] Choose Lock now. All revealed fields clear.
- [ ] Unlock again and leave the popup open for two minutes. It locks and clears the values.

## Form fill and deletion

- [ ] Start with the empty local preview form. Unlock, leave Erase saved vault selected, and click Fill preview form.
- [ ] Confirm all four fields fill correctly, the popup locks, the saved vault is erased, and no form is submitted.
- [ ] Reopen the extension and confirm there is no saved vault.
- [ ] Clear the page form and confirm all fields are blank.
- [ ] Repeat with Erase saved vault unselected. After fill, the popup locks and the encrypted vault can still be unlocked.
- [ ] Use Erase vault. Close/reopen and confirm it is gone.
- [ ] Verify filled page values clear after two minutes and when leaving/hiding the page.
- [ ] Press Enter in a populated sample field. Confirm the URL does not acquire profile query parameters.

## Rejection paths

- [ ] Attempt fill on an unrelated site. It must reject it.
- [ ] Attempt fill on a different local path or port. It must reject it.
- [ ] Attempt fill when any preview field is already populated. It must refuse to overwrite.
- [ ] Begin unlock, then close the popup immediately. Reopening must show a locked vault.
- [ ] Navigate the active tab during fill. It must not fill the new document.
- [ ] Check DevTools for extension errors. Do not export logs that contain profile values or passphrases.

## Website

- [ ] Search, use Next and Previous, filter by region, open and close a broker record, and check an empty result.
- [ ] Review both desktop and narrow layouts. Buttons retain transparent interiors, bright borders, gradient hover borders, and white text.
- [ ] Confirm the UI never suggests a directory listing proves personal exposure or completed removal.

Record Chrome version, operating system, source commit, results, and reproducible failures. Keep this record free of personal details.
