/**
 * Non-normative implementation projection of the approved Figma visual source.
 * Source file yXE7a9vAyWdbU9iLnjFmXf, production Pages 19–37 (PD-0033);
 * the implementation contract is Page 37 «Handoff & Full Coverage».
 * Values below mirror the file's local variables (5 collections, 113 variables),
 * 11 text styles and 5 effect styles read via Figma MCP on 2026-08-05.
 */
export const figmaVisualSource = {
  fileKey: "yXE7a9vAyWdbU9iLnjFmXf",
  handoffPage: "293:20",
  iaPermissionsPage: "293:3",
  patternsPage: "293:18",
  flowsPage: "293:19",
  productionPages: {
    "21-student-today": "293:4",
    "22-progress-profile": "293:5",
    "23-practice-homework": "293:6",
    "24-schedule-events": "293:7",
    "25-teachers-school": "293:8",
    "26-teacher-today": "293:9",
    "27-teacher-students": "293:10",
    "28-community": "293:11",
    "29-admin-operations": "293:12",
    "30-owner": "293:13",
    "31-notifications-activity": "293:14",
    "32-auth-account-privacy": "293:15",
    "33-system-states": "293:16",
    "34-a11y-locale-motion": "293:17",
  },
  components: {
    premiumContextHero: "300:2",
    growthSignal: "302:2",
    evidenceCard: "304:2",
    agendaEntry: "306:2",
    explainableInsight: "308:2",
    lessonRecap: "309:20119",
    bottomNavigation: "310:20542",
    dateChip: "333:231",
    rsvpControl: "333:268",
  },
} as const;

/**
 * Primitive palette — collection «Belcanto · Primitives».
 * Not for direct use in screens: components consume `semantic` roles.
 */
export const palette = {
  ink950: "#070611",
  ink900: "#0B0918",
  ink850: "#100D21",
  ink800: "#151229",
  ink750: "#1B1733",
  ink700: "#25203F",
  ink650: "#31294B",
  paper0: "#FFFFFF",
  paper50: "#FCFAFF",
  paper100: "#F7F3FB",
  paper200: "#EEE8F4",
  paper300: "#DED5E8",
  lavender200: "#E8DDFC",
  lavender300: "#D8C4FF",
  lavender400: "#B998FF",
  violet400: "#A78BFA",
  violet500: "#9C6FFF",
  violet600: "#7C3AED",
  violet700: "#6D28D9",
  violet800: "#5B21B6",
  magenta500: "#E84F92",
  gold300: "#F9DC91",
  gold400: "#F6C56B",
  gold500: "#DFA846",
  cyan400: "#67E8F9",
  green500: "#42C297",
  amber500: "#EFB457",
  red500: "#EB5B70",
  gray300: "#D2CCDC",
  gray400: "#ABA3BA",
  gray600: "#777080",
  black1000: "#000000",
  glassDark: "rgba(21,18,41,0.87)",
  glassLight: "rgba(255,255,255,0.91)",
  violetGlow: "rgba(156,111,255,0.30)",
  white08: "rgba(255,255,255,0.08)",
  white12: "rgba(255,255,255,0.12)",
  black60: "rgba(0,0,0,0.60)",
} as const;

export type ColorMode = "dark" | "light";

/**
 * Semantic color roles — collection «Belcanto · Color», both Figma modes.
 * Production UI is dark-first; the light projection ships with the tokens so
 * the aliases stay shared (Page 37, HOF-01 base contract).
 */
export const semanticByMode = {
  dark: {
    bgCanvas: palette.ink950,
    bgSunken: palette.ink900,
    bgSurface: palette.ink850,
    bgRaised: palette.ink800,
    bgGlass: palette.glassDark,
    bgAction: palette.violet500,
    bgActionPressed: palette.violet600,
    bgCommunity: palette.magenta500,
    textPrimary: palette.paper50,
    textSecondary: palette.gray300,
    textMuted: palette.gray400,
    textInverse: palette.ink950,
    textOnAction: palette.ink950,
    textAccent: palette.lavender300,
    textGold: palette.gold300,
    textCommunity: palette.magenta500,
    borderDefault: palette.ink700,
    borderStrong: palette.ink650,
    borderAccent: palette.violet500,
    borderGlass: palette.white12,
    feedbackSuccess: palette.green500,
    feedbackWarning: palette.amber500,
    feedbackDanger: palette.red500,
    accentViolet: palette.violet500,
    accentMagenta: palette.magenta500,
    accentGold: palette.gold400,
    accentCyan: palette.cyan400,
    iconDefault: palette.gray300,
    iconPrimary: palette.violet400,
    overlayScrim: palette.black60,
  },
  light: {
    bgCanvas: palette.paper100,
    bgSunken: palette.paper200,
    bgSurface: palette.paper0,
    bgRaised: palette.paper50,
    bgGlass: palette.glassLight,
    bgAction: palette.violet600,
    bgActionPressed: palette.violet700,
    bgCommunity: palette.magenta500,
    textPrimary: palette.ink950,
    textSecondary: palette.ink650,
    textMuted: palette.gray600,
    textInverse: palette.paper0,
    textOnAction: palette.paper0,
    textAccent: palette.violet700,
    textGold: palette.gold500,
    textCommunity: palette.magenta500,
    borderDefault: palette.paper300,
    borderStrong: palette.gray400,
    borderAccent: palette.violet600,
    borderGlass: palette.paper0,
    feedbackSuccess: palette.green500,
    feedbackWarning: palette.amber500,
    feedbackDanger: palette.red500,
    accentViolet: palette.violet600,
    accentMagenta: palette.magenta500,
    accentGold: palette.gold500,
    accentCyan: palette.cyan400,
    iconDefault: palette.ink650,
    iconPrimary: palette.violet600,
    overlayScrim: palette.black60,
  },
} as const;

/** Mode-agnostic role map: same keys in both projections, string values. */
export type SemanticColors = Record<keyof (typeof semanticByMode)["dark"], string>;

/**
 * Default projection used by production components. Light-first by the
 * owner's decision of 2026-08-06 («дизайн очень строгий, не хватает
 * красок»): both modes ship in the Figma token contract, the light
 * aliases are the documented counterpart — no value is invented here.
 * The dark projection stays available for a future theme switch.
 */
export const semantic: SemanticColors = semanticByMode.light;

/** Scale — collection «Belcanto · Scale». */
export const space = {
  s0: 0,
  s1: 4,
  s2: 8,
  s3: 12,
  s4: 16,
  s5: 20,
  s6: 24,
  s8: 32,
  s10: 40,
  s12: 48,
  s16: 64,
} as const;

export const radius = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 22,
  xxl: 28,
  pill: 999,
} as const;

export const sizes = {
  touchMin: 48,
  iconSm: 16,
  iconMd: 20,
  iconLg: 24,
  avatarSm: 32,
  avatarMd: 48,
  avatarLg: 72,
  /** Text-field shell heights (Page 32 field / 347:213 composer). */
  inputMin: 64,
  inputErrorMin: 82,
} as const;

/**
 * Bottom-navigation host geometry — projection of the Figma component
 * «Bottom Navigation · Production» (310:20542) and its fixed host frame
 * on every production screen (x=12, y=768, 366×68 inside the 390×844
 * viewport): the bar is pinned outside the scroll, never inside it.
 */
export const navigation = {
  height: 68,
  itemMinHeight: 56,
  sideInset: space.s3,
  bottomGap: space.s2,
  maxWidth: 366,
} as const;

export const strokes = {
  hairline: 1,
  default: 1.5,
} as const;

export const opacities = {
  disabled: 0.42,
  scrim: 0.6,
  scrimStrong: 0.68,
} as const;

/** Motion — collection «Belcanto · Motion» (HOF-15: tokens 100–320 ms). */
export const motion = {
  durationQuick: 120,
  durationStandard: 200,
  durationExpressive: 320,
  durationCelebration: 480,
  durationReduced: 0,
  distanceSubtle: 4,
  distanceStandard: 12,
  easingStandard: [0.2, 0, 0, 1] as const,
  easingExit: [0.4, 0, 1, 1] as const,
} as const;

/** Text styles — the 11 local Figma text styles, mapped to loaded Onest fonts. */
export const typeStyles = {
  displayXl: { fontFamily: "Onest_800ExtraBold", fontSize: 40, lineHeight: 48, letterSpacing: -0.8 },
  displayL: { fontFamily: "Onest_700Bold", fontSize: 32, lineHeight: 40, letterSpacing: -0.5 },
  headingL: { fontFamily: "Onest_700Bold", fontSize: 24, lineHeight: 32, letterSpacing: -0.2 },
  headingM: { fontFamily: "Onest_600SemiBold", fontSize: 20, lineHeight: 28, letterSpacing: -0.1 },
  bodyL: { fontFamily: "Onest_400Regular", fontSize: 18, lineHeight: 28, letterSpacing: 0 },
  bodyM: { fontFamily: "Onest_400Regular", fontSize: 16, lineHeight: 24, letterSpacing: 0 },
  bodyS: { fontFamily: "Onest_400Regular", fontSize: 14, lineHeight: 20, letterSpacing: 0 },
  labelL: { fontFamily: "Onest_600SemiBold", fontSize: 14, lineHeight: 20, letterSpacing: 0.1 },
  labelM: { fontFamily: "Onest_600SemiBold", fontSize: 12, lineHeight: 16, letterSpacing: 0.2 },
  caption: { fontFamily: "Onest_400Regular", fontSize: 12, lineHeight: 16, letterSpacing: 0.1 },
  metricXl: { fontFamily: "Onest_800ExtraBold", fontSize: 48, lineHeight: 52, letterSpacing: -1.2 },
} as const;

export type TypeStyleName = keyof typeof typeStyles;

/**
 * Effect styles. Figma blur values are kept verbatim in `figma`; the RN
 * projections approximate them (shadowRadius ≈ blur / 2) and give Android
 * an elevation step per level.
 */
export const elevation = {
  subtle: {
    figma: { color: "rgba(0,0,0,0.22)", offsetY: 4, blur: 12, spread: 0 },
    shadowColor: palette.black1000,
    shadowOpacity: 0.22,
    shadowOffset: { width: 0, height: 4 },
    shadowRadius: 6,
    elevation: 3,
  },
  raised: {
    figma: { color: "rgba(0,0,0,0.34)", offsetY: 10, blur: 28, spread: -4 },
    shadowColor: palette.black1000,
    shadowOpacity: 0.34,
    shadowOffset: { width: 0, height: 10 },
    shadowRadius: 14,
    elevation: 8,
  },
  overlay: {
    figma: { color: "rgba(0,0,0,0.48)", offsetY: 18, blur: 44, spread: -8 },
    shadowColor: palette.black1000,
    shadowOpacity: 0.48,
    shadowOffset: { width: 0, height: 18 },
    shadowRadius: 22,
    elevation: 16,
  },
  glowViolet: {
    figma: { color: "rgba(156,112,255,0.34)", offsetY: 4, blur: 22, spread: 0 },
    shadowColor: palette.violet500,
    shadowOpacity: 0.34,
    shadowOffset: { width: 0, height: 4 },
    shadowRadius: 11,
    elevation: 6,
  },
  glowGold: {
    figma: { color: "rgba(245,184,82,0.28)", offsetY: 4, blur: 22, spread: 0 },
    shadowColor: palette.gold400,
    shadowOpacity: 0.28,
    shadowOffset: { width: 0, height: 4 },
    shadowRadius: 11,
    elevation: 6,
  },
} as const;

/**
 * Gradient fills used by feature cards, hero overlays and the initials
 * avatar. Values ship with the approved B.0 visual layer and stay until
 * the Page 21/32 re-skin replaces the surfaces that consume them.
 */
export const gradients = {
  feature: ["#7D3BED", "#211245"] as const,
  badge: ["#9C70FF", "#5C29BD"] as const,
  welcomeOverlay: [
    "rgba(8,5,18,0.05)",
    "rgba(8,5,18,0.42)",
    "#080512",
  ] as const,
  homeOverlay: [
    "rgba(8,5,18,0.05)",
    "rgba(8,5,18,0.25)",
    "#080512",
  ] as const,
} as const;

/*
 * The legacy B.0 token layer (colors/radii/spacing/fonts/typeScale) is
 * gone: every screen consumes the semantic projection above. What
 * remains below is the B.0 auth/hero decoration geometry — image and
 * badge dimensions with no Figma Pages 19–37 counterpart. They leave
 * together with the Page-21/32 re-skin of Welcome, SignIn and the
 * student home hero.
 */
export const metrics = {
  authBadge: 76,
  authBadgeTop: 18,
  authForgotGap: 10,
  authFormGap: 36,
  authGlowHeight: 240,
  authGlowWidth: 260,
  confirmationBadge: 86,
  welcomeHeroHeight: 470,
  welcomeMinimumHeight: 740,
  homeHeroHeight: 264,
} as const;
