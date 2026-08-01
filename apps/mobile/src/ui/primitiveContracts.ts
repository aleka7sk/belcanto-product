import type { ReactNode } from "react";

export type PrimitiveState = {
  disabled?: boolean;
  busy?: boolean;
  selected?: boolean;
  expanded?: boolean;
  invalid?: boolean;
};

export interface AccessiblePrimitiveContract {
  accessibilityLabel?: string;
  accessibilityHint?: string;
  state?: PrimitiveState;
  testID?: string;
}

export interface AppSurfaceContract extends AccessiblePrimitiveContract {
  children?: ReactNode;
}

export interface ScrollFrameContract extends AccessiblePrimitiveContract {
  children?: ReactNode;
  keyboardAware?: boolean;
}

export interface StackContract extends AccessiblePrimitiveContract {
  children?: ReactNode;
  direction: "vertical" | "horizontal";
}

export interface TextContract extends AccessiblePrimitiveContract {
  children: ReactNode;
  semanticRole: "body" | "label" | "supporting";
}

export interface HeadingContract extends AccessiblePrimitiveContract {
  children: ReactNode;
  level: 1 | 2 | 3;
}

export interface PressableContract extends AccessiblePrimitiveContract {
  children?: ReactNode;
  onPress(): void;
}

export interface ButtonContract extends AccessiblePrimitiveContract {
  children: ReactNode;
  onPress(): void;
}

export interface InputContract extends AccessiblePrimitiveContract {
  value: string;
  onChangeText(value: string): void;
  inputMode?: "text" | "tel" | "email";
  secureTextEntry?: boolean;
}

export interface FormFieldContract extends AccessiblePrimitiveContract {
  id: string;
  label: ReactNode;
  input: ReactNode;
  supportingText?: ReactNode;
  error?: ReactNode;
  required?: boolean;
}

export interface SemanticPrimitiveContracts {
  AppSurface: AppSurfaceContract;
  ScrollFrame: ScrollFrameContract;
  Stack: StackContract;
  Text: TextContract;
  Heading: HeadingContract;
  Pressable: PressableContract;
  Button: ButtonContract;
  Input: InputContract;
  FormField: FormFieldContract;
}
