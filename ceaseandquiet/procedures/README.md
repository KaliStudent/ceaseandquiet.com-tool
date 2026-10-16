# Broker procedure intelligence — draft schema

The owner's authorized mailbox was reviewed for procedure *patterns*. The initial sample showed acknowledgement, extra identity or residence verification, "no record found", and processed or completed claims. These are distinct states. A reply claiming completion is not independent evidence of deletion.

`procedure.schema.json` stores only controlled categories, a broker directory ID, review dates, and (after independent verification) an official public HTTPS source. It intentionally has no fields for names, addresses, email addresses, ticket numbers, message IDs, private links, exact reply text, attachments, credentials, or case histories.

Review workflow:

1. Read owner-authorized correspondence privately. Do not export message bodies or identifying metadata into Git, logs, or a public dashboard.
2. Classify the procedure with the schema vocabulary. Mark mailbox-derived claims `needs-independent-verification`.
3. Check the broker's official public instructions separately, add an official source and date, then change status only after review.
4. Do not automate a live request from a procedure record. A procedure may change or require additional consent and identity checks.

No individual mailbox-derived records are committed in this stage. The source mailbox remains the only location of the original correspondence.
