import Link from "next/link";
import { getApiUrl } from "@/services/apiService";

export default function Navbar() {
  return (
    <div className="div flex items-center justify-between">
      <Link href="/">
        <h1 className="text-5xl font-bold text-transparent bg-gradient-to-r from-indigo-500 to-purple-800 bg-clip-text mb-2 pb-5">
          Study Hub
        </h1>
      </Link>
      <div className="flex items-center justify-around gap-4">
        <div></div>
        <a
          href={`${getApiUrl()}/api/auth/signout`}
          className="text-md underline text-gray-500"
        >
          Sign out
        </a>
        <Link href="/credits" className="text-md underline text-gray-500">
          Credits
        </Link>
      </div>
    </div>
  );
}
