# UI skills for HopVista

Researched October 17, 2026 (America/Los_Angeles). Primary sources only. This is research and a proposed workflow; no UI, package, MCP, or hook configuration was changed.

## Recommendation

**Use Impeccable first for the product's design direction and ongoing reviews. Use Motion when an interaction needs animation beyond simple CSS. Bring in Bklit when a real diagnostic question calls for a chart. Treat Kokonut as an optional source of individual components.** These tools fill different roles, so choosing one overall winner would hide the useful distinctions. This ordering is my recommendation for the current repository, rather than a vendor claim.

| Tool | Actual role | Best use in HopVista | Current adoption cost |
| --- | --- | --- | --- |
| Impeccable | Design guidance, workflows, critique, and optional detector/browser tooling | Clarify the target form, results hierarchy, observations, evidence, and empty/error states | Can guide the existing React/CSS code; richer tooling has separate setup |
| Motion | General animation library plus an agent skill/MCP toolkit | Meaningful state transitions, exit/layout animation, interruptible gestures | Add the `motion` package only when needed; no Tailwind required |
| Bklit UI | Composable chart source distributed through shadcn registries | Compare successful phase durations or show genuinely ordered measurement samples | Normal setup introduces shadcn/Tailwind infrastructure and chart dependencies |
| Kokonut UI | Ready-made React components with styling and animation | A selected navigation, search, drawer, or other interaction that fits the agreed design | Normal setup requires Tailwind v4, shadcn configuration, and per-component dependencies |

The vendors describe these roles in their [Impeccable workflow](https://impeccable.style/docs/new-work/), [Motion React guide](https://motion.dev/docs/react), [Bklit introduction](https://bklit.com/docs), and [Kokonut installation guide](https://kokonutui.com/docs).

## What is installed versus what runs

An **agent skill** supplies instructions and reference material to the coding agent. A **library or copied component** becomes application code. An **MCP server** supplies tools to the agent, such as documentation or registry search. Installing the skill files does not install npm packages, configure MCP servers, or add UI components. Motion's [installer documentation](https://motion.dev/docs/ai-kit-install) and Kokonut's [agent guide](https://kokonutui.com/docs/agents) describe those additional integration steps separately.

The installed entries are `impeccable`, `motion`, `bklit-ui`, `bklit-ui-design`, `bklit-ui-migrate`, and `install-kokonutui-component`. Their availability does not establish that every optional tool advertised on their websites works in this session. In particular, copied Impeccable skill files are not proof that its native CLI/browser tooling or automatic hooks have been initialized. Its [getting-started guide](https://impeccable.style/tutorials/getting-started/) separates project context, installation, and reviews, and its [hook guide](https://impeccable.style/docs/hooks/) requires separate activation/trust.

## Impeccable

This is the best starting point here: it helps decide what the page communicates, how users move through it, and which visual system supports that job. It also remains useful after the first draft. The command set includes `shape`, `critique`, `audit`, `polish`, `harden`, and `clarify`; these cover planning, UX evaluation, technical quality, finishing, edge cases, and copy. It is not a widget library. [Official repository](https://github.com/pbakaus/impeccable).

Current documentation distinguishes durable product context in `PRODUCT.md` from visual-system decisions in `DESIGN.md`. New work can start with generated mockups or directly in code. Mockups can explore more ambitious compositions, but their translation into code needs iteration; code-first is a sensible default for extending this existing utility. Image generation is optional. [New-work workflow](https://impeccable.style/docs/new-work/).

Use it before choosing decorative components, and use focused critique/polish passes after implementation. My recommendation is to keep diagnostic truth, readability, and evidence access as explicit constraints so stylistic preferences do not override useful existing conventions. Its automated detectors supplement visual and usability review; they do not establish accessibility compliance. [Audit capabilities](https://github.com/pbakaus/impeccable), [automatic checks](https://impeccable.style/docs/hooks/).

Impeccable's public skills, commands, CLI, and detection engine are free under Apache 2.0. Coding-agent and image-generation usage can still cost money. Its documented installer requires Node.js 18.19 or newer. No paid Impeccable purchase is needed to start. [FAQ](https://impeccable.style/faq/), [first steps](https://impeccable.style/tutorials/getting-started/).

## Motion

Your initial understanding was right: Motion is the broad animation tool. The free React library handles declarative animation, gestures, layout changes, and exits. Its own documentation recommends CSS transitions for simple isolated effects, such as a hover color change. Use Motion when state coordination, interruption, layout, or removal from the DOM makes plain CSS cumbersome. [React guide](https://motion.dev/docs/react).

The core library is MIT licensed. Installation uses `npm install motion` and imports from `motion/react`. React 18.2 or newer is supported, and Vite needs no special integration, so HopVista's React 19/Vite foundation fits. This does not require Next.js or Tailwind. [License](https://raw.githubusercontent.com/motiondivision/motion/main/LICENSE.md), [installation](https://motion.dev/docs/react-installation).

The separate AI Kit includes guidance and MCP-powered services. Current official docs say best practices and documentation search are free; CSS spring generation, MotionScore audits, transition editing, and premium example/component source require Motion+. The GitHub README's spring-generation wording is less explicit, so do not assume the generator is free merely because the skill describes it. Manual CSS animation and ordinary Motion-library development remain available without Motion+. [AI Kit](https://motion.dev/docs/ai-kit), [AI Kit installation](https://motion.dev/docs/ai-kit-install), [repository](https://github.com/motiondivision/ai-kit).

Motion UI is another separate offering: its premium runtime components use a private registry and require Motion+ credentials at install/build time. The free core package is sufficient for ordinary custom animation. [Motion UI installation](https://motion.dev/docs/react-motion-ui-installation).

Configure reduced motion deliberately. `MotionConfig reducedMotion="user"` suppresses transform/layout animation while allowing other values, including opacity, to animate. Use `useReducedMotion` for bespoke decisions such as disabling parallax or continuous effects. Favor transforms/opacity, avoid expensive layout/paint work, and profile on actual devices; vendor frame-rate descriptions are not a guarantee for this app. [Accessibility](https://motion.dev/docs/react-accessibility), [performance](https://motion.dev/docs/performance).

## Bklit UI

This is the strongest specialist among these four for charts. Its public skills cover chart installation/composition (`bklit-ui`), visual styling (`bklit-ui-design`), and Recharts/shadcn chart migration (`bklit-ui-migrate`). HopVista has no existing Recharts charts, so the migration skill currently has no job. [Skills guide](https://bklit.com/docs/skills).

Bklit installs source through the shadcn CLI; the normal documented path first sets up shadcn and then adds selected `@bklit` items. In an existing Vite application, that setup includes Tailwind and aliases. This is meaningful infrastructure work for HopVista's present plain-CSS UI. It does not require replacing React/Vite with Next.js. Hand-porting source to ordinary CSS is possible engineering work, but is outside the standard installation path. [Bklit installation](https://bklit.com/docs/installation), [shadcn Vite setup](https://ui.shadcn.com/docs/installation/vite).

The chart source is MIT licensed; Bklit Studio's implementation is proprietary. The hosted Studio can be used for its intended purpose, and the repository presents it as a playground for tuning/exporting chart compositions. Do not copy Studio's application source into HopVista under the chart license. [Repository license distinction](https://github.com/bklit/bklit-ui), [Studio license](https://raw.githubusercontent.com/bklit/bklit-ui/main/LICENSE-STUDIO.md).

Chart theming uses CSS variables, which can be mapped to the product's eventual tokens. Read the actual installed source before coding: some documentation examples use `@bklitui/ui/charts`, whereas the consumer skill documents local imports from files written to `components/charts/`. Do not assume an internal example import is an npm dependency to install. [Theming](https://bklit.com/docs/theming), [consumer imports](https://bklit.com/docs/skills).

Reduced-motion support needs special attention. The installed skill's `rules/animation.md` explicitly describes partial coverage: bar grow/stagger, several chart entrances, tooltip springs, and live scrolling need additional handling. It recommends gating series animation with `animate={!reduce}` and warns that `animationDuration={0}` alone is insufficient for bars. A web-cached older copy of the same upstream file claims broader coverage, so the installed version and actual downloaded chart implementation must be checked together. [Installed animation guidance](C:/Users/natha/.codex/skills/bklit-ui/rules/animation.md), [older upstream file surfaced by search](https://raw.githubusercontent.com/bklit/bklit-ui/main/skills/bklit-ui/rules/animation.md).

Bklit publishes device-local benchmarks and explains nonidentical library configurations. These are useful experiments, not proof of universal superiority. Its documented default Cartesian entrance takes about 1100 ms, with data updates also animated. Keep chart identity stable and avoid replaying entrances on every HopVista polling response. Validate keyboard access, screen-reader meaning, and a text/table equivalent before shipping rather than relying on hover tooltips alone. The latter checks are my implementation recommendations. [Benchmark methodology](https://bklit.com/benchmarks), [bar API](https://bklit.com/docs/components/bar-chart).

## Kokonut UI

Kokonut supplies finished component ideas rather than a competing general animation engine. Many components already use Motion, so it often complements Motion rather than replacing it. Its installation guide requires Tailwind CSS v4; shadcn configuration is needed for CLI installation, while hand-copying still leaves dependency installation to you. A selected component can bring icons, UI primitives, and Motion along with its source. [Installation](https://kokonutui.com/docs).

For HopVista, evaluate it only after deciding the interaction needed. A drawer may help on small screens; an ornamental animated background does little to explain a network diagnosis. That is a product-fit judgment. The documented Smooth Drawer, for example, combines Vaul, React, Motion, and Tailwind, illustrating why a demo is more than an animation snippet. [Smooth Drawer](https://kokonutui.com/docs/navigation/smooth-drawer).

The open-source repository is MIT licensed. Kokonut UI Pro is a separate paid catalog with a private registry and API key; its terms permit commercial projects but prohibit redistributing/reselling the components or making a component library from them. Free skill installation does not grant Pro access. Avoid a purchase until a specific component has demonstrated value. [Open-source repository](https://github.com/kokonut-labs/kokonutui), [Pro offering and usage terms](https://kokonutui.pro/).

The pages inspected do not establish a blanket accessibility or reduced-motion guarantee across the catalog. Review the selected component's keyboard/focus behavior, semantics, touch behavior, reduced-motion path, and cost of continuous effects. These are acceptance criteria, not an assertion that every Kokonut component fails them.

## Proposed HopVista workflow

Repository observations: `ui/package.json` currently lists React 19.2, TypeScript, and Vite, with no Tailwind, shadcn, Motion, or chart package. `ui/src/main.tsx` uses ordinary CSS, polls active runs every 400 ms, distinguishes completion/timeouts/skipped outcomes, and links observations to evidence. The README describes a Go executable embedding the built frontend. Those local files are the primary sources for the following recommendations.

1. **Clarify the product and preserve diagnostic truth.** Use Impeccable context/shape work to establish the audience, main task, constraints, and result hierarchy. Review existing product/domain documents before generating new context files. Explain direct-connection scope, incomplete evidence, and IPv4 scope accurately.
2. **Design one useful screen.** Improve the target form, running/stopping/completed states, observations, measurement details, and evidence navigation. Agree on typography, spacing, colors, and responsive behavior. Start in the current React/plain-CSS stack; a component registry is not required to design well.
3. **Decide whether charts answer a real question.** A bar chart can compare completed DNS/TCP/TLS/HTTP-first-byte durations, with explicit units and separate timeout/skipped states. Never turn a timeout into zero or connect unrelated protocol categories as a time series. Only use a time-series chart for genuinely ordered samples, and preserve the distinction between observed RTT samples and link delay. A simple CSS/SVG visualization may be sufficient before adopting Bklit infrastructure.
4. **Select optional components.** If a Kokonut component solves an identified interaction better than the current controls, inspect that one component and estimate the styling/dependency work. If Bklit is justified, establish shadcn/Tailwind once inside `ui`, preserve the current stylesheet deliberately, install only the needed chart, and map theme tokens.
5. **Add purposeful motion.** Use CSS for uncomplicated feedback. Use Motion for state/layout/exit coordination when justified. Animate meaningful transitions instead of each 400 ms data refresh. Respect reduced motion and keep pending/failed states truthful.
6. **Review the working result.** Use Impeccable critique, audit, harden, and polish at the appropriate stage. Verify narrow screens, keyboard/focus, text zoom, screen-reader labels, missing data, cancellations, long addresses, and stable live updates. Run the frontend build and the repository's applicable checks after implementation, then verify the embedded Go-served output. No runtime Node service is inherently required by these client-side libraries; extra agent/dev tooling stays a development concern.

## Scoped prompts to use later

These are suggested prompts, not commands executed during this research.

```text
$impeccable shape the HopVista diagnostic results screen. Read the current
UI and domain docs first. Prioritize observations and evidence, preserve
measurement caveats and every outcome, and propose directions before coding.
Keep React/Vite and plain CSS for the first iteration.
```

```text
$motion review the proposed running-to-completed and evidence-expansion
transitions. Use CSS where adequate and identify where Motion would help.
Respect reduced motion and avoid animation on every polling update.
```

```text
$bklit-ui evaluate one bar chart comparing completed exchange phase
durations. Explain setup changes first. Keep timeouts and skipped outcomes
explicit, preserve the text measurements, and check reduced-motion behavior.
```

```text
$install-kokonutui-component investigate one drawer component for narrow
screen evidence details. Inspect its source, dependencies, accessibility,
and fit with our tokens before proposing installation changes.
```

## Limits of this investigation

This research verified official docs and inspected the installed guidance and current frontend foundation. It did not run skill launchers, purchase premium access, initialize MCP servers/hooks, install application dependencies, execute vendor benchmarks, or test components in the browser. Performance and accessibility therefore remain implementation checks, and advertised optional tooling should not be described as configured or verified merely because its skill exists.
