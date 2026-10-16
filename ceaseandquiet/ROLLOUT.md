# Local beta rollout

## 1. Development candidate

Build on the review branch, run the dedicated tests and CI, and package the preview executable, extension, license, review packet, and this checklist. Synthetic details only. Native Chrome verification is required to finish this stage.

## 2. Final review

The owner chooses the security tools/team. Give every reviewer the same source commit, SECURITY-REVIEW.md, and BETA-CHECKLIST.md. Preserve independent findings before comparing them. Resolve reproducible findings and review the updated commit.

## 3. Private beta

After review and the Chrome checklist pass, the owner can approve a small beta group. Document permitted data, device assumptions, deletion limitations, and support instructions. The beta remains a local form-fill tool; filling does not create a server request.

## 4. Future public service

Public hosting, real personal data intake, authenticated accounts, verified broker procedures, request authorization, and outbound delivery require their own implementation and review. Shared cPanel/Podman services and DNS changes remain subject to the owner's existing review requirement. No production deployment, DNS change, or mail delivery is part of this candidate.
