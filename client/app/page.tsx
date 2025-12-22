"use client";

import { useState, useEffect } from "react";
import TicketCatalog from "@/components/TicketCatalog";
import BookingForm from "@/components/BookingForm";
import { getAvailability } from "@/lib/api";
import { TierInfo, Tier, getAvailableCount } from "@/types";

export default function Home() {
  const [tiers, setTiers] = useState<TierInfo[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedTier, setSelectedTier] = useState<Tier | null>(null);
  const [selectedTierPrice, setSelectedTierPrice] = useState<number>(0);
  const [selectedTierInfo, setSelectedTierInfo] = useState<
    TierInfo | undefined
  >(undefined);

  const fetchAvailability = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getAvailability();
      if (response.success && response.tiers) {
        setTiers(response.tiers);
      } else {
        setError("Failed to load ticket availability");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load availability"
      );
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAvailability();
  }, []);

  // Update selectedTierInfo when tiers or selectedTier changes
  useEffect(() => {
    if (selectedTier && tiers.length > 0) {
      const tierInfo = tiers.find((t) => t.tier === selectedTier);
      if (tierInfo) {
        setSelectedTierInfo(tierInfo);
      }
    }
  }, [selectedTier, tiers]);

  const handleTierSelect = (tier: Tier, price: number) => {
    setSelectedTier(tier);
    setSelectedTierPrice(price);
    // Find and set the tier info with availableList
    const tierInfo = tiers.find((t) => t.tier === tier);
    setSelectedTierInfo(tierInfo);
  };

  const handleBookingSuccess = async () => {
    // Refresh availability after successful booking
    await fetchAvailability();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <header className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Concert Ticket Booking
          </h1>
          <p className="text-gray-600">
            Book your tickets for the upcoming concert. Select a tier to get
            started.
          </p>
        </header>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800">{error}</p>
            <button
              onClick={fetchAvailability}
              className="mt-2 text-sm text-red-700 hover:text-red-900 underline"
            >
              Retry
            </button>
          </div>
        )}

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4 text-gray-900">
            Available Tickets
          </h2>
          <TicketCatalog tiers={tiers} isLoading={isLoading} />
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4 text-gray-900">
            Select Ticket Tier
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {tiers.map((tierInfo) => (
              <button
                key={tierInfo.tier}
                onClick={() => handleTierSelect(tierInfo.tier, tierInfo.price)}
                disabled={
                  getAvailableCount(
                    tierInfo.totalSeats,
                    tierInfo.reservedCount
                  ) === 0
                }
                className={`p-4 border-2 rounded-lg text-left transition-all ${
                  selectedTier === tierInfo.tier
                    ? "border-blue-500 bg-blue-50"
                    : "border-gray-200 bg-white hover:border-gray-300"
                } ${
                  getAvailableCount(
                    tierInfo.totalSeats,
                    tierInfo.reservedCount
                  ) === 0
                    ? "opacity-50 cursor-not-allowed"
                    : "cursor-pointer"
                }`}
              >
                <div className="font-semibold text-gray-500 text-lg mb-1">
                  {tierInfo.tier === "VIP"
                    ? "VIP"
                    : tierInfo.tier === "FRONT_ROW"
                    ? "Front Row"
                    : "General Admission"}
                </div>
                <div className="text-sm text-gray-600">
                  {getAvailableCount(
                    tierInfo.totalSeats,
                    tierInfo.reservedCount
                  )}{" "}
                  available
                </div>
              </button>
            ))}
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4 text-gray-900">
            Book Your Ticket
          </h2>
          <BookingForm
            selectedTier={selectedTier}
            tierPrice={selectedTierPrice}
            tierInfo={selectedTierInfo}
            onBookingSuccess={handleBookingSuccess}
          />
        </section>
      </div>
    </div>
  );
}
