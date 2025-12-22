"use client";

import {
  TierInfo,
  formatPrice,
  getTierDisplayName,
  getAvailableCount,
} from "@/types";

interface TicketCatalogProps {
  tiers: TierInfo[];
  isLoading?: boolean;
}

export default function TicketCatalog({
  tiers,
  isLoading,
}: TicketCatalogProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="border rounded-lg p-6 bg-gray-50 animate-pulse"
          >
            <div className="h-6 bg-gray-200 rounded w-1/2 mb-4"></div>
            <div className="h-8 bg-gray-200 rounded w-1/3 mb-4"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2"></div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      {tiers.map((tierInfo) => (
        <div
          key={tierInfo.tier}
          className="border rounded-lg p-6 bg-white shadow-sm hover:shadow-md transition-shadow"
        >
          <div className="flex justify-between items-start mb-4">
            <h3 className="text-xl font-semibold text-gray-900">
              {getTierDisplayName(tierInfo.tier)}
            </h3>
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium ${
                getAvailableCount(tierInfo.totalSeats, tierInfo.reservedCount) >
                0
                  ? "bg-green-100 text-green-800"
                  : "bg-red-100 text-red-800"
              }`}
            >
              {getAvailableCount(tierInfo.totalSeats, tierInfo.reservedCount) >
              0
                ? "Available"
                : "Sold Out"}
            </span>
          </div>

          <div className="mb-4">
            <p className="text-3xl font-bold text-gray-900">
              {formatPrice(tierInfo.price)}
            </p>
            <p className="text-sm text-gray-500 mt-1">per ticket</p>
          </div>

          <div className="pt-4 border-t">
            <p className="text-sm text-gray-600">
              <span className="font-medium">
                {getAvailableCount(tierInfo.totalSeats, tierInfo.reservedCount)}
              </span>{" "}
              tickets available
            </p>
          </div>
        </div>
      ))}
    </div>
  );
}
