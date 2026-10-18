# HopVista

## Platform

web

## Users

Curious non-specialists investigating why a website is slow from their computer. Technical evidence is one step away.

## Product Purpose

Help a person understand connection delays for one target URL through a diagnostic run, recorded measurements, and an observed finding with explicit limits. Use the terms in GLOSSARY.md.

## Capabilities and Constraints

The approved specification is GitHub issue #10: https://github.com/DarrylN2/hopvista/issues/10. The first release measures a direct connection to one selected IPv4 address from a local Windows 11 x64 executable. It does not measure browser rendering or establish a root cause. React/TypeScript, Vite, and plain CSS present the Go-owned results and findings.

A 60-second measurement cutoff stops scheduling work; native work already started may finish later. Keep stopping visible until cleanup completes. Preserve partial measurements and distinct outcomes. One active diagnostic run and the latest terminal run result live in memory; process exit clears them.

## Brand Commitments

Name: HopVista. Voice: clear, measured, approachable, explicit about uncertainty. The user approved Clear field notes with guided reading and its crisp graphic refinement: light neutral canvas, bold ink typography, confident blue actions and path accents, selective monospace measurements, generous spacing and ruled sections.

## Evidence on Hand

Issue #10 and docs/implementation/v1-acceptance-status.md define evidence requirements and open acceptance work. The initial tested platform condition is Windows 11 25H2 build 26200 x64 with an unelevated token; broader platform coverage is unverified. Distribution remains experimental.

## Accessibility & Inclusion

Provide keyboard access, text status labels, visible focus, contextual evidence navigation, narrow-screen reflow, and reduced-motion support. Human comprehension and rendered accessibility review remain required for release.
