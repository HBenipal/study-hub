"use client";

import React, { useEffect } from "react";
import { useState } from "react";
import { getApiUrl, signIn, signUp } from "@/services/apiService";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { TextInput } from "@/components/TextInput";
import { PrimaryButton } from "@/components/PrimaryButton";
import { Divider } from "@/components/Divider";
import Link from "next/link";
import "../App.css";

export default function Home() {
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [formType, setFormType] = useState<"signup" | "signin">("signin");

  const router = useRouter();

  const handleSignIn = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    try {
      await signIn(username, password);
      router.push("/");
    } catch (error) {
      let errorMsg = "";
      if (error instanceof Error) {
        errorMsg = ": " + error.message;
      }
      toast.error(`error signing in${errorMsg}`, {
        closeButton: true,
      });
    }
  };

  const handleSignUp = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    try {
      await signUp(username, password);
      router.push("/");
    } catch (error) {
      let errorMsg = "";
      if (error instanceof Error) {
        errorMsg = ": " + error.message;
      }
      toast.error(`error signing up${errorMsg}`, {
        closeButton: true,
      });
    }
  };

  const handleFormSubmit = async (e?: React.FormEvent) => {
    switch (formType) {
      case "signin":
        handleSignIn(e);
        break;
      case "signup":
        handleSignUp(e);
        break;
    }
  };

  const getFormSubmitButtonLabel = (): string => {
    switch (formType) {
      case "signin":
        return "Sign in";
      case "signup":
        return "Sign up";
    }
  };

  const toggleFormState = () => {
    switch (formType) {
      case "signin":
        setFormType("signup");
        break;
      case "signup":
        setFormType("signin");
        break;
    }
  };

  const GITHUB_CLIENT_ID = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID;

  const [githubAuthUrl, setGithubAuthUrl] = useState<string>("");

  useEffect(() => {
    if (typeof window !== "undefined") {
      const REDIRECT_URI = `${window.location.origin}/auth/github_redirect`;
      setGithubAuthUrl(
        `https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${REDIRECT_URI}&scope=read:user%20repo`,
      );
    }
  }, []);

  return (
    <div className="flex w-screen h-screen overflow-hidden">
      <div className="w-1/2 h-full border-r border-gray-300 flex items-center justify-center bg-gradient-to-bl from-indigo-500 to-purple-800">
        <div className="w-fit flex flex-col items-start justify-center">
          <h1 className="text-8xl font-bold text-white">Study Hub</h1>

          <h1 className="text-3xl font-bold text-white typewriter">
            Collaborate in real-time
          </h1>
        </div>
      </div>
      <div className="flex items-center justify-center flex-col w-1/2 h-full">
        <form
          onSubmit={handleFormSubmit}
          className="flex gap-6 items-center justify-center flex-col w-full"
        >
          <TextInput
            value={username}
            placeholder="Username"
            onChange={setUsername}
            className="w-8/12"
          />
          <TextInput
            type="password"
            value={password}
            placeholder="Password"
            onChange={setPassword}
            className="w-8/12"
          />
          <PrimaryButton
            buttonType="submit"
            label={getFormSubmitButtonLabel()}
            isDisabled={false}
            className="w-8/12"
          />
        </form>

        <Divider />

        <Link href={githubAuthUrl} className="w-8/12">
          <button className="w-full bg-zinc-800 text-white font-semibold px-4 py-2 rounded">
            Continue with GitHub
          </button>
        </Link>

        <div className="flex items-center justify-center gap-2 mt-10 text-gray-500 text-md">
          <p>{formType == "signin" ? "New user?" : "Existing user?"}</p>
          <p
            className="underline font-semibold cursor-pointer"
            onClick={() => toggleFormState()}
          >
            {formType == "signin" ? "Sign up" : "Sign in"}
          </p>
        </div>
      </div>
    </div>
  );
}
