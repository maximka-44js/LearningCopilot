import {
  openAccountConsole,
  signInWithKeycloak,
  signOutWithKeycloak,
} from "@/features/auth/actions";

type AuthLinksProps = {
  isAuthenticated: boolean;
};

export function AuthLinks({ isAuthenticated }: AuthLinksProps) {
  if (isAuthenticated) {
    return (
      <div className="flex gap-12">
        <form action={openAccountConsole}>
          <button className="button" type="submit">
            Account
          </button>
        </form>
        <form action={signOutWithKeycloak}>
          <button className="button" type="submit">
            Sign out
          </button>
        </form>
      </div>
    );
  }

  return (
    <div className="flex gap-12">
      <form action={signInWithKeycloak}>
        <button className="button" type="submit">
          Sign in
        </button>
      </form>
    </div>
  );
}
