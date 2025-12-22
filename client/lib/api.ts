import { AvailabilityResponse, BookingOrder, BookingResponse } from "@/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Generate idempotency key (client-side UUID)
function generateIdempotencyKey(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`;
}

// Get or create idempotency key from sessionStorage
export function getIdempotencyKey(bookingKey: string): string {
  if (typeof window === "undefined") {
    return generateIdempotencyKey();
  }

  const storageKey = `idempotency_${bookingKey}`;
  let key = sessionStorage.getItem(storageKey);

  if (!key) {
    key = generateIdempotencyKey();
    sessionStorage.setItem(storageKey, key);
  }

  return key;
}

// Clear idempotency key after successful booking
export function clearIdempotencyKey(bookingKey: string): void {
  if (typeof window === "undefined") return;
  const storageKey = `idempotency_${bookingKey}`;
  sessionStorage.removeItem(storageKey);
}

export async function getAvailability(): Promise<AvailabilityResponse> {
  const response = await fetch(`${API_BASE_URL}/booking/availability`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  console.log("GET availability response", response);

  if (!response.ok) {
    throw new Error(`Failed to fetch availability: ${response.statusText}`);
  }

  return response.json();
}

export async function bookTicket(
  order: Omit<
    BookingOrder,
    | "idempotencyKey"
    | "totalAmtInUSCent"
    | "status"
    | "paymentID"
    | "paymentStatus"
  >,
  idempotencyKey: string
): Promise<BookingResponse> {
  // Simulate payment - generate a payment ID and set status to CONFIRMED
  const paymentID = `pay_${Date.now()}_${Math.random()
    .toString(36)
    .substring(2, 15)}`;
  const paymentStatus = "CONFIRMED" as const;
  const status = "PENDING" as const;

  const bookingOrder: BookingOrder = {
    ...order,
    idempotencyKey,
    status,
    paymentID,
    paymentStatus,
  };

  const response = await fetch(`${API_BASE_URL}/booking/ticket`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(bookingOrder),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.message || `Booking failed: ${response.statusText}`
    );
  }

  return response.json();
}
