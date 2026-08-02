/**
 * Non-normative implementation projection of the approved Figma visual source.
 * Source: yXE7a9vAyWdbU9iLnjFmXf, Auth 4:4 and Student 4:5.
 * PD-0033 keeps Figma authoritative until Design/Technical review is complete.
 */
export const figmaVisualSource = {
  fileKey: "yXE7a9vAyWdbU9iLnjFmXf",
  authPage: "4:4",
  studentPage: "4:5",
  welcome: "26:33",
  signIn: "26:35",
  confirmation: "26:37",
  studentHome: "30:3",
} as const;

export const colors = {
  canvas: "#070611",
  surface: "#100D21",
  raised: "#151229",
  raisedGlass: "rgba(21,18,41,0.87)",
  border: "#25203F",
  borderGlass: "rgba(255,255,255,0.12)",
  textPrimary: "#FCFAFF",
  textSecondary: "#D2CCDC",
  textMuted: "#ABA3BA",
  textAccent: "#D8C4FF",
  textGold: "#F9DC91",
  textOnAction: "#070611",
  violet: "#9C6FFF",
  violetPressed: "#7C3AED",
  violetDeep: "#211245",
  cyan: "#67E8F9",
  gold: "#F6C56B",
  danger: "#EB5B70",
  success: "#67E8F9",
  transparent: "rgba(0,0,0,0)",
  scrim: "rgba(7,6,17,0.72)",
} as const;

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

export const radii = {
  input: 12,
  control: 14,
  compactCard: 18,
  card: 20,
  feature: 22,
  pill: 999,
} as const;

export const spacing = {
  xxs: 2,
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  field: 18,
  xl: 20,
  xxl: 24,
  section: 30,
  loose: 40,
} as const;

export const metrics = {
  contentMaxWidth: 430,
  authGutter: 20,
  homeGutter: 18,
  minimumTarget: 48,
  inputMinHeight: 64,
  errorInputMinHeight: 82,
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
  workflowEyebrowTop: 74,
  borderWidth: 1,
} as const;

export const fonts = {
  regular: "Onest_400Regular",
  medium: "Onest_500Medium",
  semibold: "Onest_600SemiBold",
  bold: "Onest_700Bold",
  extrabold: "Onest_800ExtraBold",
} as const;

export const typeScale = {
  brand: { fontSize: 13, letterSpacing: 2.4, lineHeight: 17 },
  homeBrand: { fontSize: 11, letterSpacing: 2, lineHeight: 15 },
  welcomeTitle: { fontSize: 34, lineHeight: 41 },
  authTitle: { fontSize: 30, lineHeight: 36 },
  screenTitle: { fontSize: 28, lineHeight: 34 },
  homeName: { fontSize: 27, lineHeight: 32 },
  cardTitle: { fontSize: 19, lineHeight: 23 },
  sectionTitle: { fontSize: 16, lineHeight: 21 },
  body: { fontSize: 14, lineHeight: 20 },
  bodyLarge: { fontSize: 15, lineHeight: 23 },
  supporting: { fontSize: 12, lineHeight: 17 },
  label: { fontSize: 11, lineHeight: 15 },
  eyebrow: { fontSize: 10, lineHeight: 13, letterSpacing: 1 },
  micro: { fontSize: 8, lineHeight: 11, letterSpacing: 0.8 },
} as const;
