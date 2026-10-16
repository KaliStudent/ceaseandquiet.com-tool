# Native Chrome verification record

Date: 2026-10-16. Windows, Chrome 155.0.8059.84, existing owner profile.

Tested executable source: 52a52396a8536bae14d8ab75db0779ef9096a660.
Extension: local unpacked Manifest V3 version 0.2.0. The owner performed installation. The local packaged preview served http://127.0.0.1:8787/.

Only synthetic profile values and a generated test passphrase were used. No real personal details or passphrases are included in this record.

## Observed passes

| Check | Observation |
| --- | --- |
| Save | Encrypt and save cleared revealed fields and locked the popup. |
| Reopen | Saved vault remained locked; profile values were hidden. |
| Wrong passphrase | Unlock was rejected, popup stayed locked, and passphrase cleared. |
| Correct unlock | Stored synthetic profile values were revealed correctly. |
| Default fill | All four sample values transferred to the local form; popup locked and reported saved vault erased. Reopening showed no saved vault. |
| Retain option | With erase-after-fill unselected, fill locked the popup and the saved vault could be unlocked again. |
| Two-minute lock | An undisturbed repeat showed the locked state and timeout message after 129 seconds, with no revealed profile fields. |
| Final cleanup | A final default fill reported saved vault erased. Clear form then left the page's sample fields blank. |

The first timer attempt was inconclusive because the owner interacted with the popup; only the undisturbed repeat is counted. A later attempted occupied-form check happened after the form had cleared and therefore established another successful fill, not overwrite rejection.

The successful native fill exercises Chrome's actual storage, activeTab/scripting, and document-ID injection flow. It does not establish every browser permission or race behavior.

## Remaining acceptance work

- Repeat the full BETA-CHECKLIST.md in a clean development profile.
- Verify explicit Lock now and Erase vault actions, occupied-form refusal, unsupported origins/paths/ports, immediate-close-during-unlock, and navigation during fill in native Chrome.
- Complete the native page expiry/hide and Enter-key submission checks, and inspect extension errors without exporting sensitive logs.
- Obtain the owner's final security review using the chosen tools/team and reconcile findings against one source commit.

Automated WebCrypto, session, popup, and destination tests cover several remaining failure paths with simulated Chrome APIs. Website checks also passed in the in-app browser, including directory interactions and a 390-pixel viewport. This record is a development verification report and does not claim a security audit or private-beta approval.
