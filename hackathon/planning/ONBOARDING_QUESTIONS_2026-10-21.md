# Questions for the onboarding chat, 21 October 2026

Onboarding: 21 October, 12:00–13:30 UTC (YouTube, GenAI Academy). Written by the advisor at 10:07 UTC, for the owner to
paste into the chat. Each question is one line, ready to paste. The ones marked **MUST** decide how we build.

Already answered, so not asked again: pre-existing components are allowed (the organisers confirmed); the Cursor agent
CLI counts as building in Cursor.

## MUST

1. **Meterless, local or cloud?** The Setup Guide says Meterless Nightshift must run locally as the primary AI provider,
   but the Academy lesson shows Meterless as a cloud provider. Which one counts for eligibility?
2. **"Running on your machine".** Does Nightshift in a virtual machine on our own Mac count? And Zetaris on a server in
   our own network rather than on the laptop?
3. **Nightshift installers.** The macOS installers are not signed or notarized, so macOS refuses them. Is a notarized
   build, or a Linux build, coming?
4. **Sponsor access.** When do participants get the free Zetaris, Meterless and NVIDIA accounts or credits?

## SHOULD

5. **Zetaris setup route.** The Academy lesson uses Freemium on port 9001; the Setup Guide uses the platform bundle on
   port 3000. Which one is the reference for judging?
6. **Zetaris REST tables.** Is there a way to send a nested JSON body in a POST (needed for OSV's `/v1/query`), or to pass
   URL parameters at query time instead of fixing them when the table is created?
7. **Zetaris and files.** Can Zetaris query JSON or CSV files placed in its own storage? What is the recommended way to
   load a large dataset, such as FIRST's full EPSS CSV?
8. **Meterless Engines.** Does our own Go implementation of H-MEM and World Model count as implementing an engine, or
   must the official package be used?
9. **How judges run entries.** Will judges run projects themselves? If so, do they have Zetaris and Nightshift, or should
   we ship a self-contained lab (Docker Compose)?

## NICE TO KNOW

10. **Submission metadata.** Is there an official format for declaring which sponsor technologies an entry uses? We have
    a draft `metadata.json` that other teams could share.
11. **Declaring prior art.** Where in the submission should pre-existing components be declared?
12. **Proof of timing.** Is the repository's push history enough to show the work was done in the window, or do you want
    something else?
