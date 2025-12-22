// Types matching the Go backend models

export type Tier = "VIP" | "FRONT_ROW" | "GA";

export type BookingStatus = "PENDING" | "CONFIRMED" | "FAILED" | "CANCELED";

export type PaymentStatus = "PENDING" | "CONFIRMED" | "FAILED" | "CANCELED";

export interface Booking {
  id: string;
  userId: string;
  tier: Tier;
  status: BookingStatus;
  idempotencyKey?: string;
  seatNo: number;
  country: string;
  zipCode: string;
  currency: string;
  totalAmtInUSCent: number;
  paymentID: string;
  paymentStatus: PaymentStatus;
  createdAt: string;
  updatedAt: string;
}

export interface BookingOrder {
  userId: string;
  tier: Tier;
  status: BookingStatus;
  idempotencyKey: string;
  country: string;
  zipCode: string;
  currency: string;
  seatNo: number;
  // totalAmtInUSCent: number; this will be calculated on the server
  paymentID: string;
  paymentStatus: PaymentStatus;
}

export interface BookingResponse {
  success: boolean;
  message?: string;
  booking?: Booking;
}

export interface TierInfo {
  tier: Tier;
  price: number; // in US cents
  totalSeats: number; // total seats for this tier (from server)
  reservedCount: number; // number of seats reserved for this tier
  availableList?: number[]; // AvailableList from server
}

// Helper to calculate available count from totalSeats and reservedCount
export function getAvailableCount(
  totalSeats: number,
  reservedCount: number
): number {
  return totalSeats - reservedCount;
}

export interface AvailabilityResponse {
  success: boolean;
  message?: string;
  reservedSeats?: Record<string, number[]>;
  tiers?: TierInfo[];
}

// Helper to get tier display name
export function getTierDisplayName(tier: Tier): string {
  switch (tier) {
    case "VIP":
      return "VIP";
    case "FRONT_ROW":
      return "Front Row";
    case "GA":
      return "General Admission";
    default:
      return tier;
  }
}

// Helper to format price in USD
export function formatPrice(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

// Helper to get seat range for tier
export function getSeatRangeForTier(tier: Tier): { min: number; max: number } {
  switch (tier) {
    case "VIP":
      return { min: 1, max: 30 };
    case "FRONT_ROW":
      return { min: 31, max: 60 };
    case "GA":
      return { min: 61, max: 100 };
    default:
      return { min: 1, max: 100 };
  }
}
