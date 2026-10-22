---
name: HopVista
description: Clear field notes with guided reading for local connection diagnostics.
colors:
  canvas: "#f7f8f5"
  paper: "#fff"
  ink: "#24352f"
  muted: "#58665f"
  rule: "#dce2db"
  wash: "#eef2ed"
  blue: "#244fcd"
  blue-hover: "#193da8"
  blue-active: "#122e82"
  error: "#923a30"
  error-wash: "#fcf0ed"
  disabled-background: "#e4e8e2"
  disabled-text: "#667168"
  secondary-border: "#a3afa6"
  input-border: "#819187"
  placeholder: "#6a756e"
  state-background: "#e8ede7"
  running-text: "#205477"
  running-background: "#e8f0f7"
  stopping-text: "#745022"
  stopping-background: "#f5eddf"
  blue-wash: "#edf1fc"
  blue-ink: "#1c3154"
  blue-muted: "#495c83"
  teal: "#126960"
  teal-wash: "#e6f3f0"
  teal-rule: "#a9d6cb"
  teal-muted: "#416b66"
  amber: "#745022"
  amber-wash: "#f5eddf"
  selection-arrival: "#dce7ff"
  copy-border: "#b8c7ec"
  path-node-border: "#244fcd"
  path-connector: "#8b9b90"
  workspace-border: "#9aa79e"
  finding-divider: "#c7d1eb"
typography:
  headline:
    fontFamily: '"Segoe UI", system-ui, sans-serif'
    fontSize: "3rem"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "-.035em"
  title:
    fontSize: "1.4rem"
    fontWeight: 700
    lineHeight: 1.35
  section:
    fontSize: "1.25rem"
    fontWeight: 700
    letterSpacing: "-.02em"
  body:
    fontFamily: '"Segoe UI", system-ui, sans-serif'
    lineHeight: 1.55
  label:
    fontSize: ".875rem"
    fontWeight: 600
  observation:
    fontSize: "1.125rem"
    fontWeight: 700
    lineHeight: 1.5
  data:
    fontFamily: '"Cascadia Mono", Consolas, "SFMono-Regular", monospace'
  measured-value:
    fontFamily: '"Cascadia Mono", Consolas, "SFMono-Regular", monospace'
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: "-.025em"
rounded:
  panel: "12px"
  control: "6px"
  badge: "4px"
  evidence-link: "2px"
spacing:
  compact: "8px"
  control-gap: "12px"
  row: "16px"
  inset: "20px"
  panel: "24px"
  section: "28px"
  stack: "32px"
  column: "40px"
  page: "48px"
components:
  button-primary:
    backgroundColor: "{colors.blue}"
    textColor: "{colors.paper}"
    rounded: "{rounded.control}"
    padding: ".65rem 1.1rem"
  button-primary-hover:
    backgroundColor: "{colors.blue-hover}"
  button-primary-active:
    backgroundColor: "{colors.blue-active}"
  button-primary-disabled:
    backgroundColor: "{colors.disabled-background}"
    textColor: "{colors.disabled-text}"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: ".65rem 1.1rem"
  button-copy:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.blue}"
    rounded: "{rounded.badge}"
    padding: "4px 8px"
  input:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: ".6rem .8rem"
  finding:
    backgroundColor: "{colors.blue-wash}"
    rounded: "0 0 12px 12px"
    padding: "24px"
  evidence-card:
    backgroundColor: "{colors.paper}"
    rounded: "{rounded.panel}"
    padding: "20px"
  state:
    backgroundColor: "{colors.state-background}"
    textColor: "{colors.ink}"
    rounded: "{rounded.badge}"
    padding: "5px 10px"
---

# Design System: HopVista

## Overview

**Creative North Star: "Clear field notes with guided reading"**

A light neutral canvas, ink typography, confident blue actions, generous spacing, and ruled sections make diagnostic observations approachable to a curious non-specialist. The approved crisp graphic treatment strengthens the existing field-notes direction: a compact wordmark, bold balanced headings, blue hop markers, and selective monospace addresses and measurements. Familiar Segoe UI body copy keeps explanations easy to read. The composition still guides attention from an observed finding and its limits to separate website timings and the recorded path.

This system is extracted from `ui/src/style.css` and `ui/src/main.tsx`. The user approved this bolder typography, clearer explanations, and polish pass while preserving the composition and measurement semantics. Go-owned observation and limitation strings remain authoritative; supplementary explanations do not reinterpret their evidence.

The earlier system received representative live and visual verification, with actual-run captures at `.impeccable/review/desktop.png` and `.impeccable/review/mobile.png`. For this graphic pass an independent reviewer inspected synthetic completed states at 1440px and 390px, stopping at 390px, partial at 1440px, and failed at 390px, returning **ship at representative scope** with no material fixes. Current captures are under `.impeccable/review/graphic/`; captures beyond that reviewed set do not imply visual acceptance. Seventeen current state/width checks found no horizontal overflow, including long 64-hop content at 320px; pending active states and evidence focus/return were also verified. This is not whole-release acceptance or a full human, screen-reader, or zoom audit. The type detector was attempted but failed because its engine was missing; no clean detector result is claimed.

**Key Characteristics:**

- Light neutral surfaces with dark ink and muted supporting text.
- Confident blue controls and visible evidence links.
- Flat panels, strong section rules, and generous space between reading groups.
- Bold sans headings and selective monospace readings, with observations paired with limits and evidence.
- Path on the left and finding, limits, and timings on the right; summary, timing, then path on narrow screens.

## Colors

### Primary

The confident blue action family identifies primary buttons, the second headline line, path glyph and hop markers, finding accents, evidence links, text selection, caret, and keyboard focus. Hover and active shades deepen this same family. The blue wash supports the finding and evidence-link hover; the finding has a 3px blue top rule.

Teal identifies the website-timing category: its heading and top rule, phase labels, introductory explanation and conceptual phase tags. It does not encode a fast or successful measurement. Blue-tinted ink and secondary text keep the finding wash cohesive. Selected evidence receives the same blue wash; a short brighter wash acknowledges arrival. These colors supplement explicit headings, labels and focus indicators.

### Neutral

Canvas is the page background; paper supports the diagnostic form and evidence records. Ink carries headings, labels, and readings; muted carries explanations and caveats. Rule separates reading groups; wash distinguishes raw evidence blocks without elevation.

### Status

Running uses the blue status pair. Stopping and cancelled use the warm status pair. Failed and request errors use the error pair. Other run states use the neutral state background. Status always includes words; color does not carry the meaning alone.

## Typography

The body family remains Segoe UI with system-ui and sans-serif fallbacks. No custom font asset or separate display face is established. Body size inherits browser defaults; do not infer a fixed body size token. Headings are bold and balanced; the second line of the intro uses blue. The fixed intro scale is 3rem at weight 700 and line-height 1.1, reducing to 2.25rem at 760px. Once a run starts, it compacts to 1.75rem and then 1.5rem at 480px. The run target heading is 1.4rem, reducing to 1.125rem at 760px.

The data stack is Cascadia Mono, Consolas, SFMono-Regular, monospace. Use it selectively for selected/reply/next-hop addresses, raw RTT values, hop numbers, completed timing values, technical identifiers, raw JSON, and the conceptual empty-state phase labels. It is not the body family. Completed timing values use 1.25rem at weight 600, line-height 1.35, and -.025em tracking; unavailable outcomes retain the body face. Timing, hops, echo readings, and evidence fields use tabular numerals.

Section headings use 1.25rem at weight 700 and -.02em tracking; fourth-level headings use 1rem at weight 700. Smaller context commonly uses .8125rem; raw JSON uses .75rem/1.6 in the data family. Introductory copy has a maximum measure of 58ch, the empty state 65ch, and technical explanatory copy 75ch. Long target URLs, addresses, errors, and evidence content wrap rather than forcing horizontal page overflow. Panels also use anywhere wrapping to protect long finding URLs.

## Layout

The centered page container has a maximum width of (1280px), including its horizontal padding. Page gutters are (48px), becoming (28px) at (960px) and (20px) at (760px). The masthead separates identity from local diagnostic scope with a 2px ink bottom rule. The form wraps its target field, advanced disclosure, and actions; on narrow screens these stack at full width.

For the implemented run result, the desktop grid uses (1.25fr 1fr), a (40px) column gap, and an additional (40px) right inset on the path column before its dividing rule. The path is left; the summary column holds the finding and limitations above website timing on the right. At (960px), gap and path inset reduce to (24px), finding padding becomes (20px), hop labels stack, and timing labels narrow from (110px) to (80px).

At (760px), the result becomes one column in DOM reading order: finding and limitations, website timing, then path. The path gains a top rule; technical evidence becomes a single column. The desktop evidence grid has two equal columns with a (24px) gap. At (480px), the conceptual empty-state guide stacks; its path and phase illustrations align at 53px and contain no measured values. At (400px), echo address and RTT readings stack. These are observed layouts of the diagnostic surface, not a requirement that every future screen use this composition.

## Elevation & Depth

The current stylesheet defines no shadows. Depth comes from neutral surface changes, fine borders, and spacing. The finding uses blue wash and a blue top rule; evidence records use paper with a rule border. Keyboard focus is a (2px) blue outline with a (4px) offset, including programmatically focused evidence records.

## Shapes

Panels use the shared rounded panel token; controls and raw evidence blocks use the control radius. Status badges and text-like evidence controls have smaller corners. Borders are generally (1px), with (2px) ink section rules and the (3px) finding accent. The finding has square top corners and rounded bottom corners. The path uses circular nodes and dotted vertical connectors; those connectors mean probe order only and do not establish physical links or the HTTP route.

## Components

### Buttons

Primary actions are blue, semibold, and at least (44px) high. Hover deepens the blue, active deepens it again, and disabled controls use neutral colors with a not-allowed cursor. The secondary cancel action is transparent with a neutral border and ink text; hover uses wash. All interactive controls expose the shared keyboard focus treatment.

Start/cancel and copy button colors change immediately so enabled and acknowledged states retain text contrast. Other button backgrounds transition in 160ms. The authored motion moment is evidence selection: the focused record receives a 520ms wash acknowledgment and stays visibly selected. Evidence arrows move 3px on hover or keyboard focus over 160ms. The running status dot, newest hop marker and successful-copy glyph each have a single 220–240ms scale/opacity acknowledgment using cubic-bezier(.16, 1, .3, 1). No animation loops, staggered page entrances, fake progress or altered measurement values are used. Reduced motion removes animations, transitions and arrow movement, retaining static selection, live copy feedback and automatic scrolling.

The compact Copy IPv4 action appears only for a selected address, writes that literal address to the clipboard and acknowledges success with a blue Copied label and check glyph for 2400ms. Its width stays at least 104px to avoid shifting on acknowledgment. A polite status announces success; a visible fallback explains manual copying when clipboard access fails. Pending and copied feedback describe the action, not network health. The button is at least 32px high, or 44px for coarse pointers.

The accessibility follow-up retains focus during pending copy requests with
aria-disabled/busy feedback. If copying fails, a labelled read-only field appears
with the literal address focused and selected for keyboard copying. Start/cancel
requests focus the persistent status message with the shared outline, preserving
focus when action controls disable or disappear. The selected scope includes an
assistive-text label without changing its visible styling. See the scoped
[issue #16 evidence](docs/implementation/issue-16.md); human review remains open.

### Inputs and disclosures

Inputs have paper fill, a neutral stroke, a blue caret, and a minimum height of (46px). Hover darkens the border to ink. Visible labels precede the URL and hop count fields. Advanced options use a native details disclosure, and summaries have a minimum height of (44px). The hop count field caps its width at (116px).

### Identity and navigation

The masthead combines the HopVista wordmark with its existing inline path glyph and a right-aligned scope label. The glyph is (36px) square with a 2px blue stroke; the wordmark uses 1.5rem, weight 700, and -.04em tracking. A keyboard skip link appears on focus and leads to the diagnostic workspace. There is no implemented application menu to extrapolate.

### Finding and timing

The finding panel groups selected IPv4, observation, limitation, and evidence references. Active runs without a selected IPv4 say **Pending**; terminal runs without one say **No IPv4 address selected**. The observed finding stays pending until the run finishes. Static finding-type explanations supplement the unchanged Go observation and limitation. A native optional disclosure defines the dominant-wait rule for applicable findings: at least two comparable completed phases in one exchange, a longest wait of at least 100 ms, and at least twice the next longest. Timing uses separate exchange blocks and ruled definition-list rows for DNS, TCP, TLS, and HTTP first byte, each with explanatory text and evidence access. Completed values receive the data typography; incomplete outcomes are never rendered as zero.

### Status badges

Compact, softly rounded badges pair state text with the status colors described above. The ready/running notice uses a small dot alongside text in a polite live region. Request failures use an alert message in the error wash.

### Path and evidence

Hop rows pair hop number, observed reply information, raw RTT where present, and an evidence button. Dotted connectors express probe order. Evidence controls pair underlined text with an inline arrow, with a minimum height of (32px), increasing to (44px) for coarse pointers.

Finding references use phase names when available, such as “View TCP evidence”. Main visual labels omit raw IDs; accessible names and titles retain them, and technical evidence shows them. Technical evidence uses a native disclosure. Selecting a measurement's evidence opens it, focuses and scrolls to the matching record, and marks the record with a blue border. A return control restores focus to the originating measurement. Raw JSON is separately disclosed, wraps, and has a maximum height of (28rem) with internal scrolling.

## Do's and Don'ts

### Do:

- Do keep the canvas light, text in ink, and blue used for actions, evidence access, identity, hop markers, and finding emphasis.
- Do keep monospace selective to data and technical roles, with Segoe UI explanations.
- Do pair an observed finding with its limitation and recorded evidence.
- Do preserve visible labels, keyboard focus, and textual status meaning.
- Do preserve the summary, timing, then path reading order on narrow screens.
- Do label unavailable and pending measurements explicitly; missing durations and RTTs are not zero.

### Don't:

- Don't interpret dotted probe-order connectors as physical network links or the HTTP route.
- Don't present an active run's missing IPv4 selection as a terminal failure; show Pending.
- Don't present diagnostic observations as a root cause or website timings as page load measurements.
- Don't treat limited completed-run checks as acceptance of all states, edge cases, or the whole release.
