"use client";

import { useState, useEffect } from "react";
import {
  Tier,
  getSeatRangeForTier,
  formatPrice,
  getTierDisplayNameAndColor,
  TierInfo,
} from "@/types";
import {
  bookTicket,
  getIdempotencyKey,
  clearIdempotencyKey,
  getAvailability,
} from "@/lib/api";

// Generate UUID v4
function generateUUID(): string {
  if (typeof window !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // Fallback for browsers that don't support crypto.randomUUID
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

interface BookingFormProps {
  selectedTier: Tier | null;
  tierPrice: number;
  tierInfo?: TierInfo; // Tier info with availableList from server
  onBookingSuccess: () => void;
}

export default function BookingForm({
  selectedTier,
  tierPrice,
  tierInfo,
  onBookingSuccess,
}: BookingFormProps) {
  // Auto-generate userId UUID on component mount
  const [userId] = useState<string>(() => generateUUID());

  const [formData, setFormData] = useState({
    country: "",
    zipCode: "",
    currency: "USD",
    seatNo: "",
  });
  const [availableSeats, setAvailableSeats] = useState<number[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [bookingId, setBookingId] = useState<string | null>(null);

  // Use availableList from tierInfo when tier is selected
  useEffect(() => {
    if (!selectedTier || !tierInfo) {
      setAvailableSeats([]);
      setFormData((prev) => ({ ...prev, seatNo: "" }));
      return;
    }

    // Use availableList directly from server
    const available = tierInfo.availableList || [];
    setAvailableSeats(available);

    // Reset seat selection if current seat is no longer available
    setFormData((prev) => {
      const currentSeat = parseInt(prev.seatNo) || 0;
      if (!available.includes(currentSeat)) {
        return { ...prev, seatNo: "" };
      }
      return prev;
    });
  }, [selectedTier, tierInfo]);

  if (!selectedTier) {
    return (
      <div className="border rounded-lg p-6 bg-gray-50 text-center text-gray-500">
        Please select a ticket tier to book
      </div>
    );
  }

  const seatRange = getSeatRangeForTier(selectedTier);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);
    setIsSubmitting(true);

    try {
      // Validate seat number
      const seatNo = parseInt(formData.seatNo, 10);
      if (isNaN(seatNo) || seatNo < seatRange.min || seatNo > seatRange.max) {
        throw new Error(
          `Seat number must be between ${seatRange.min} and ${
            seatRange.max
          } for ${getTierDisplayNameAndColor(selectedTier).name}`
        );
      }

      // Generate or retrieve idempotency key
      const bookingKey = `${selectedTier}-${seatNo}-${userId}`;
      const idempotencyKey = getIdempotencyKey(bookingKey);

      // Book ticket
      const response = await bookTicket(
        {
          userId: userId,
          tier: selectedTier,
          country: formData.country,
          zipCode: formData.zipCode,
          currency: formData.currency,
          seatNo,
        },
        idempotencyKey
      );

      if (response.success && response.booking) {
        setSuccess(true);
        setBookingId(response.booking.id);
        clearIdempotencyKey(bookingKey);
        onBookingSuccess();

        // Reset form (keep userId as it's auto-generated)
        setFormData({
          country: "",
          zipCode: "",
          currency: "USD",
          seatNo: "",
        });
        // Refresh available seats after successful booking
        // The parent component will refresh tierInfo via onBookingSuccess callback
        const seatsResponse = await getAvailability();
        if (seatsResponse.success && seatsResponse.tiers) {
          const updatedTierInfo = seatsResponse.tiers.find(
            (t) => t.tier === selectedTier
          );
          if (updatedTierInfo && updatedTierInfo.availableList) {
            setAvailableSeats(updatedTierInfo.availableList);
          }
        }
      } else {
        throw new Error(response.message || "Booking failed");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  if (success && bookingId) {
    return (
      <div className="border rounded-lg p-6 bg-green-50 border-green-200">
        <div className="flex items-center mb-2">
          <svg
            className="w-6 h-6 text-green-600 mr-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <h3 className="text-lg font-semibold text-green-900">
            Booking Successful!
          </h3>
        </div>
        <p className="text-green-800 mb-2">
          Your ticket has been booked successfully.
        </p>
        <p className="text-sm text-green-700">
          Booking ID: <span className="font-mono">{bookingId}</span>
        </p>
        <button
          onClick={() => {
            setSuccess(false);
            setBookingId(null);
          }}
          className="mt-4 text-sm text-green-700 hover:text-green-900 underline"
        >
          Book another ticket
        </button>
      </div>
    );
  }

  return (
    <div className="border rounded-lg p-6 bg-white">
      <h2 className="text-xl font-semibold text-gray-600 mb-4">
        Book {getTierDisplayNameAndColor(selectedTier).name} Ticket
      </h2>
      <p className="text-gray-600 mb-6">
        Price:{" "}
        <span className="font-bold text-lg">{formatPrice(tierPrice)}</span>
      </p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="userId"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            User ID <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="userId"
            name="userId"
            value={userId}
            readOnly
            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-600 cursor-not-allowed"
            placeholder={userId}
          />
          <p className="text-xs text-gray-500 mt-1">
            User ID is automatically generated
          </p>
        </div>

        <div>
          <label
            htmlFor="seatNo"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Seat Number <span className="text-red-500">*</span>
          </label>
          {availableSeats.length === 0 ? (
            <div className="w-full px-3 py-2 border border-red-300 rounded-md bg-red-50 text-red-600">
              No seats available for this tier
            </div>
          ) : (
            <select
              id="seatNo"
              name="seatNo"
              value={formData.seatNo}
              onChange={handleChange}
              required
              className="w-full px-3 text-gray-800 text-medium py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Select a seat</option>
              {availableSeats.map((seat) => (
                <option key={seat} value={seat}>
                  Seat {seat}
                </option>
              ))}
            </select>
          )}
          <p className="text-xs text-gray-500 mt-1">
            {availableSeats.length > 0
              ? `${availableSeats.length} seats available (${seatRange.min} - ${seatRange.max})`
              : `Valid seat range: ${seatRange.min} - ${seatRange.max}`}
          </p>
        </div>

        <div>
          <label
            htmlFor="country"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Country <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="country"
            name="country"
            value={formData.country}
            onChange={handleChange}
            required
            className="w-full px-3 py-2 border text-gray-800 text-medium border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g., USA, UK, Canada"
          />
        </div>

        <div>
          <label
            htmlFor="zipCode"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Zip Code <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="zipCode"
            name="zipCode"
            value={formData.zipCode}
            onChange={handleChange}
            required
            className="w-full px-3 py-2 text-gray-800 text-medium border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g., 10001"
          />
        </div>

        <div>
          <label
            htmlFor="currency"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Currency
          </label>
          <select
            id="currency"
            name="currency"
            value={formData.currency}
            onChange={handleChange}
            className="w-full px-3 py-2 text-gray-800 text-medium border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
            <option value="GBP">GBP</option>
            <option value="CAD">CAD</option>
            <option value="AUD">AUD</option>
          </select>
          <p className="text-xs text-gray-500 mt-1">
            Prices are displayed in USD. Currency selection is for reference
            only.
          </p>
        </div>

        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {isSubmitting
            ? "Booking..."
            : `Book Ticket - ${formatPrice(tierPrice)}`}
        </button>
      </form>
    </div>
  );
}
