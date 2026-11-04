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
        <CardTitle>Добро пожаловать</CardTitle>
        <CardDescription>
          Войдите, чтобы продолжить работу.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <form action={signInWithKeycloak}>
          <Button type="submit" className="w-full">
            Продолжить
          </Button>
        </form>
        <p className="text-xs text-slate-500">
          Продолжая, вы соглашаетесь с условиями использования и политикой
          конфиденциальности.
        </p>
      </CardContent>
    </Card>
  );
}
