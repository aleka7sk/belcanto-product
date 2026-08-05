import Svg, { Path } from "react-native-svg";

import { semantic, sizes } from "./tokens";

/**
 * Navigation and control icons exported from the Figma internal icon
 * components (__Icon/Home 12:10, Calendar 12:14, Mic 12:18, Users 12:23,
 * Trophy 12:26, Check 12:29, More 131:101). Path data is the exact exported
 * asset content (committed under assets/icons/); only the export's page
 * background artifacts were stripped. Do not edit path data by hand.
 */

export type IconName =
  | "home"
  | "calendar"
  | "mic"
  | "users"
  | "trophy"
  | "check"
  | "more";

export type IconProps = {
  size?: number;
  color?: string;
};

type StrokeIconProps = IconProps & { paths: readonly string[] };

function StrokeIcon({ paths, size = sizes.iconMd, color = semantic.iconDefault }: StrokeIconProps) {
  return (
    <Svg fill="none" height={size} viewBox="0 0 24 24" width={size}>
      {paths.map((d) => (
        <Path
          d={d}
          key={d}
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={1.8}
        />
      ))}
    </Svg>
  );
}

const HOME_PATHS = [
  "M3 11.5L12 4L21 11.5V19.5C21 19.7652 20.8946 20.0196 20.7071 20.2071C20.5196 20.3946 20.2652 20.5 20 20.5H15V14.5H9V20.5H4C3.73478 20.5 3.48043 20.3946 3.29289 20.2071C3.10536 20.0196 3 19.7652 3 19.5V11.5Z",
] as const;

const CALENDAR_PATHS = [
  "M18 5H6C4.34315 5 3 6.34315 3 8V18C3 19.6569 4.34315 21 6 21H18C19.6569 21 21 19.6569 21 18V8C21 6.34315 19.6569 5 18 5Z",
  "M7 3V7M17 3V7M3 10H21",
] as const;

const MIC_PATHS = [
  "M16 7C16 4.79086 14.2091 3 12 3C9.79086 3 8 4.79086 8 7V12C8 14.2091 9.79086 16 12 16C14.2091 16 16 14.2091 16 12V7Z",
  "M19 12C19 13.8565 18.2625 15.637 16.9497 16.9497C15.637 18.2625 13.8565 19 12 19C10.1435 19 8.36301 18.2625 7.05025 16.9497C5.7375 15.637 5 13.8565 5 12M12 19V22",
] as const;

const USERS_PATHS = [
  "M9 11C10.6569 11 12 9.65685 12 8C12 6.34315 10.6569 5 9 5C7.34315 5 6 6.34315 6 8C6 9.65685 7.34315 11 9 11Z",
  "M17 11.5C18.3807 11.5 19.5 10.3807 19.5 9C19.5 7.61929 18.3807 6.5 17 6.5C15.6193 6.5 14.5 7.61929 14.5 9C14.5 10.3807 15.6193 11.5 17 11.5Z",
  "M3 20C3.7 16 6 14 9 14C12 14 14.3 16 15 20M15 15C17.8 15 19.7 16.7 20.5 19.5",
] as const;

const TROPHY_PATHS = [
  "M12 13C13.0609 13 14.0783 12.5786 14.8284 11.8284C15.5786 11.0783 16 10.0609 16 9V4H8V9C8 10.0609 8.42143 11.0783 9.17157 11.8284C9.92172 12.5786 10.9391 13 12 13ZM12 13V18M8 6H4V8C4 9.06087 4.42143 10.0783 5.17157 10.8284C5.92172 11.5786 6.93913 12 8 12M16 6H20V8C20 9.06087 19.5786 10.0783 18.8284 10.8284C18.0783 11.5786 17.0609 12 16 12M8 21H16M9 18H15",
] as const;

const CHECK_PATHS = ["M5 12L9 16L19 6"] as const;

const MORE_PATHS = [
  "M5 13.5C5.82843 13.5 6.5 12.8284 6.5 12C6.5 11.1716 5.82843 10.5 5 10.5C4.17157 10.5 3.5 11.1716 3.5 12C3.5 12.8284 4.17157 13.5 5 13.5Z",
  "M12 13.5C12.8284 13.5 13.5 12.8284 13.5 12C13.5 11.1716 12.8284 10.5 12 10.5C11.1716 10.5 10.5 11.1716 10.5 12C10.5 12.8284 11.1716 13.5 12 13.5Z",
  "M19 13.5C19.8284 13.5 20.5 12.8284 20.5 12C20.5 11.1716 19.8284 10.5 19 10.5C18.1716 10.5 17.5 11.1716 17.5 12C17.5 12.8284 18.1716 13.5 19 13.5Z",
] as const;

export function HomeIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={HOME_PATHS} />;
}

export function CalendarIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={CALENDAR_PATHS} />;
}

export function MicIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={MIC_PATHS} />;
}

export function UsersIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={USERS_PATHS} />;
}

export function TrophyIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={TROPHY_PATHS} />;
}

export function CheckIcon(props: IconProps) {
  return <StrokeIcon {...props} paths={CHECK_PATHS} />;
}

export function MoreIcon({ size = sizes.iconMd, color = semantic.iconDefault }: IconProps) {
  return (
    <Svg fill="none" height={size} viewBox="0 0 24 24" width={size}>
      {MORE_PATHS.map((d) => (
        <Path d={d} fill={color} key={d} />
      ))}
    </Svg>
  );
}

export function Icon({ name, ...rest }: IconProps & { name: IconName }) {
  switch (name) {
    case "home":
      return <HomeIcon {...rest} />;
    case "calendar":
      return <CalendarIcon {...rest} />;
    case "mic":
      return <MicIcon {...rest} />;
    case "users":
      return <UsersIcon {...rest} />;
    case "trophy":
      return <TrophyIcon {...rest} />;
    case "check":
      return <CheckIcon {...rest} />;
    case "more":
      return <MoreIcon {...rest} />;
  }
}
