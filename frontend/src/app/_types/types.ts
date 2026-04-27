export type PaymentMethod = "cc" | "ach" | "crypto" | "venmo";
export type DonationStatus = "new" | "pending" | "success" | "failure";

export interface Donation {
  uuid: string; // unique identifier
  amount: number; // in cents, e.g. 5000 = $50.00
  currency: "USD";
  paymentMethod: PaymentMethod;
  nonprofitId: string;
  donorId: string;
  status: DonationStatus;
  createdAt: string; // ISO datetime string
  updatedAt: string; // ISO datetime string
}