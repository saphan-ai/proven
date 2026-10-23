# PV-014 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in change
- the finding that must be PROVEN: when server A answers the POST with a 307 (and with a 308) to server B under another
  host name, B receives a body that contains the synthetic token.
- evidence required: the check lines for 307 and 308, showing the token at B under change and nothing at B under base;
  the 302 line, showing B receives no body; the no-redirect line, showing B receives nothing; and change's own
  structural tests passing, which shows they do not cover the HTTP path.
- what must NOT be claimed: that the structural guard or the destination check is wrong — both work (reviewer A's
  measurements were accurate within its assigned subject); or that the vendor itself redirects.
- how the record settled it: the orchestrating agent confirmed the finding from the code and the redirect rules and
  recorded a reject. The repair gave the sender its own client that refuses redirects to other hosts, with an explicit
  proxy policy and a cross-host 307 negative control; it merged in the third round. The label rests on the redirect
  rules and the later repair, not on which reviewer rejected.
