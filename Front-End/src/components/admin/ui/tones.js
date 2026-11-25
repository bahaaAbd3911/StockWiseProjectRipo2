// Named color tones shared by badges and stat cards.
export const TONES = {
  green: "#22c55e",
  blue: "#3b82f6",
  orange: "#f59e0b",
  red: "#ef4444",
  purple: "#8b5cf6",
  gray: "#94a3b8",
};

export const toneColor = (tone) => TONES[tone] ?? tone;
