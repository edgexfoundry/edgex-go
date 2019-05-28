###############
Security Issues
###############



This page describes how to report EdgeX Foundry security issues and how they are handled. 

======================
Security Announcements
======================
Join the edgexfoundry-announce group at: https://groups.google.com/d/forum/edgexfoundry-announce) 
for emails about security and major API announcements.

=======================
Vulnerability Reporting
=======================

The EdgeX Foundry Open Source Community is grateful for all security reports made by users and security researchers. 
All reports are thoroughly investigated by a set of community volunteers.

.. _security_issue_template: https://github.com/edgexfoundry/edgex-go/blob/master/.github/ISSUE_TEMPLATE/4-security-issue-disclosure.md.
To make a report, please email the private list: security-issues@edgexfoundry.org, providing as much detail as possible.
Use the security issue template:  `security_issue_template`_.

At this time we do not yet offer an encrypted bug reporting option. 


When to Report a Vulnerability?
=====================================

- You think you discovered a potential security vulnerability in EdgeX Foundry
- You are unsure how a vulnerability affects EdgeX Foundry
- You think you discovered a vulnerability in another project that EdgeX Foundry depends upon (e.g. docker, MongoDB, Redis,..)

When NOT to Report a Vulnerability?
=========================================

- You need help tuning EdgeX Foundry components for security
- You need help applying security related updates
- Your issue is not security related

===============================
Security Vulnerability Response
===============================

Each report is acknowledged and analyzed by Product Security Committee members within one week. 

Any vulnerability information shared with Product Security Committee stays within the 
EdgeX Foundry project and will not be disseminated to other projects unless it is necessary 
to get the issue fixed.

As the security issue moves from triage, to identified fix, to release planning we will keep the reporter updated.

In the case of 3 rd party dependency (code or library not managed and maintained by the EdgeX community) 
related security issues, while the issue report triggers the same response workflow, the EdgeX community will defer to
owning community for fixes. 

On receipt of a security issue report, the security team does:

1. Discusses the issue privately to understand it

2. Uses the `Common Vulnerability Scoring System <https://www.first.org/cvss/user-guide>`_ to grade the issue

3. Determines sub-projects and developers to involve

4. Develops a fix

5. In conjunction with the product group determines when to release the fix

6. Communicates the fix

7. Uploads a Common Vulnerabilities and Exposures (CVE) style report of the issue and associated threat.

The issue reporter will be kept in the loop as appropriate. 

========================
Public Disclosure Timing
========================

A public disclosure date is negotiated by the EdgeX Product Security Committee and the bug submitter. 
We prefer to fully disclose the bug as soon as possible AFTER a mitigation is available. 
It is reasonable to delay disclosure when the bug or the fix is not yet fully understood, 
the solution is not well-tested, or for vendor coordination. The timeframe for disclosure 
may be immediate (especially publicly known issues) to a few weeks. 
The EdgeX Foundry Product Security Committee holds the final say when setting a disclosure date.

