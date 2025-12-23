"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { signInWithGithub } from "@/services/apiService";

export default function Home() {
  const [status, setStatus] = useState<"loading" | "error" | "success">(
    "loading",
  );
  const router = useRouter();
  const hasRun = useRef(false);

  const redirect = async (code: string) => {
    try {
      const response = await signInWithGithub(code);

      if (response.ok) {
        setStatus("success");
        router.push("/");
      } else {
        setStatus("error");
      }
    } catch {
      setStatus("error");
    }
  };

  useEffect(() => {
    if (hasRun.current) return; // only run once
    const params = new URLSearchParams(window.location.search);
    const code = params.get("code");
    if (!code) {
      setStatus("error");
      return;
    }

    hasRun.current = true;
    redirect(code);
  }, []);

  return (
    <div className="w-screen flex items-center justify-center">
      <h1 className="text-xl font-bold text-indigo-950 mt-4">
        {status == "error" ? "There was an error" : ""}
        {status == "loading" ? "Authenticating with GitHub..." : ""}
        {status == "success" ? "Authenticated! Redirecting..." : ""}
      </h1>
    </div>
  );
}
