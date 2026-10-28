import { NextResponse } from "next/server";

import { auth } from "@/auth";
import {
  buildSignInUrl,
  isProtectedPath,
} from "@/features/auth/lib/route-protection";

export default auth((request) => {
  const { pathname, search } = request.nextUrl;

  if (!isProtectedPath(pathname)) {
    return NextResponse.next();
  }

  if (request.auth?.user && !request.auth?.error) {
    return NextResponse.next();
  }

  const callbackUrl = `${pathname}${search}`;
  const signInUrl = buildSignInUrl(callbackUrl);

  return NextResponse.redirect(new URL(signInUrl, request.nextUrl.origin));
});

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|.*\\..*).*)"],
};