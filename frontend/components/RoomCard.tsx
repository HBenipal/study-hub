import { Room } from "@/types";
import Link from "next/link";
import { PrimaryButton } from "./PrimaryButton";

interface RoomCardProps {
  room: Room;
}

export function RoomCard({ room }: RoomCardProps) {
  return (
    <div>
      <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300 p-6 h-full">
        <div className="mb-4">
          <h3 className="text-2xl font-bold text-indigo-900 mb-2">
            {room.name}
          </h3>
          <p className="text-sm text-indigo-600 font-mono bg-purple-50 px-3 py-1 rounded inline-block">
            Code: {room.code}
          </p>
        </div>

        <div className="space-y-3 text-indigo-800">
          <div className="flex items-center justify-between">
            <span className="font-semibold">Active Users:</span>
            <span className="bg-green-200 text-green-800 px-3 py-1 rounded-full font-bold">
              {room.activeUsers}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="font-semibold block mb-1">Last Updated:</span>
            <span className="text-sm text-indigo-800">{room.lastUpdated}</span>
          </div>
        </div>

        <div className="mt-6 pt-4 border-t border-gray-200">
          <Link key={room.code} href={`/room/${room.code}`}>
            <PrimaryButton
              buttonType="button"
              label="Enter Room"
              isDisabled={false}
              className="w-full"
            />
          </Link>
        </div>
      </div>
    </div>
  );
}
