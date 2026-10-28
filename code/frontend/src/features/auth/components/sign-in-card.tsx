import { signInWithKeycloak } from "@/features/auth/actions";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function SignInCard() {
  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Welcome back</CardTitle>
        <CardDescription>
          Sign in to continue with your account.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <form action={signInWithKeycloak}>
          <Button type="submit" className="w-full">
            Continue
          </Button>
        </form>
        <p className="text-xs text-slate-500">
          By continuing, you agree to our Terms and Privacy Policy.
        </p>
      </CardContent>
    </Card>
  );
}
