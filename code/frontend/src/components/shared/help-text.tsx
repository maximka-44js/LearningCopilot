import { cn } from "@/lib/utils";

type HelpTextProps = {
  label: string;
  className?: string;
};

export default function HelpText({ label, className }: HelpTextProps) {
  return <span className={cn("leading-5", className)}>{label}</span>;
}
