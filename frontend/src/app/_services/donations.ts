import { api, type ApiResponse } from "@/app/_lib/api";
import type { Donation, DonationStatus } from "@/app/_types/types";

interface DonationsResponse {
  donations: Donation[];
}

export async function listDonations(): Promise<ApiResponse<DonationsResponse>> {
  return api.get<DonationsResponse>("/donations");
}

export async function updateDonationStatus(
  donationUUID: string,
  status: DonationStatus
): Promise<ApiResponse<Donation>> {
  return api.patch<Donation>(`/donations/${donationUUID}/status`, { status });
}
