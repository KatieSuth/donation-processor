"use client";

import { useEffect, useMemo, useState } from "react";

import { listDonations, updateDonationStatus } from "@/app/_services/donations";
import type { Donation, DonationStatus } from "@/app/_types/types";

interface Props {}
type SortColumn = "uuid" | "amount" | "nonprofitId" | "createdAt" | "updatedAt";
type SortDirection = "asc" | "desc";

const PAGE_SIZE = 10;

const statusTransitions: Record<DonationStatus, DonationStatus[]> = {
  new: ["pending"],
  pending: ["success", "failure"],
  success: [],
  failure: [],
};

function toDollars(amountInCents: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
  }).format(amountInCents / 100);
}

function toLocalDateTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "N/A";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(parsed);
}

export default function DonationsTable(_props: Props) {
  const [donations, setDonations] = useState<Donation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pendingUUID, setPendingUUID] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [sortColumn, setSortColumn] = useState<SortColumn>("createdAt");
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [currentPage, setCurrentPage] = useState(1);

  useEffect(() => {
    let isMounted = true;

    async function loadDonations() {
      setIsLoading(true);
      setError(null);

      const res = await listDonations();
      if (!isMounted) return;

      if (res.error) {
        setError(res.error);
        setDonations([]);
        setIsLoading(false);
        return;
      }

      setDonations(res.data?.donations ?? []);
      setIsLoading(false);
    }

    loadDonations();
    return () => {
      isMounted = false;
    };
  }, []);

  const sortedDonations = useMemo(() => {
    const sorted = [...donations];
    sorted.sort((a, b) => {
      let comparison = 0;
      if (sortColumn === "uuid") {
        comparison = a.uuid.localeCompare(b.uuid);
      } else if (sortColumn === "amount") {
        comparison = a.amount - b.amount;
      } else if (sortColumn === "nonprofitId") {
        comparison = a.nonprofitId.localeCompare(b.nonprofitId);
      } else if (sortColumn === "createdAt") {
        comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
      } else if (sortColumn === "updatedAt") {
        comparison = new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime();
      }
      return sortDirection === "asc" ? comparison : -comparison;
    });
    return sorted;
  }, [donations, sortColumn, sortDirection]);

  const totalPages = Math.max(1, Math.ceil(sortedDonations.length / PAGE_SIZE));
  const effectiveCurrentPage = Math.min(currentPage, totalPages);

  const paginatedDonations = useMemo(() => {
    const start = (effectiveCurrentPage - 1) * PAGE_SIZE;
    const end = start + PAGE_SIZE;
    return sortedDonations.slice(start, end);
  }, [effectiveCurrentPage, sortedDonations]);

  const paginationStart =
    sortedDonations.length === 0 ? 0 : (effectiveCurrentPage - 1) * PAGE_SIZE + 1;
  const paginationEnd = Math.min(effectiveCurrentPage * PAGE_SIZE, sortedDonations.length);

  function onSort(column: SortColumn) {
    if (sortColumn === column) {
      setSortDirection((direction) => (direction === "asc" ? "desc" : "asc"));
    } else {
      setSortColumn(column);
      setSortDirection(column === "createdAt" || column === "updatedAt" ? "desc" : "asc");
    }
    setCurrentPage(1);
  }

  function sortLabel(column: SortColumn): string {
    if (sortColumn !== column) return "Sort";
    return sortDirection === "asc" ? "ASC" : "DESC";
  }

  async function onStatusChange(donation: Donation, nextStatus: DonationStatus) {
    setPendingUUID(donation.uuid);
    setActionError(null);

    const res = await updateDonationStatus(donation.uuid, nextStatus);
    if (res.error) {
      setActionError(`Failed to update ${donation.uuid}: ${res.error}`);
      setPendingUUID(null);
      return;
    }

    const updated = res.data;
    if (!updated) {
      setActionError(`Failed to update ${donation.uuid}: empty response from API`);
      setPendingUUID(null);
      return;
    }

    setDonations((current) =>
      current.map((item) => (item.uuid === updated.uuid ? updated : item))
    );
    setPendingUUID(null);
  }

  if (isLoading) {
    return (
      <div className="rounded-lg border border-slate-700 bg-slate-950 p-4 text-sm text-slate-300">
        Loading donations...
      </div>
    );
  }

  return (
    <section className="rounded-lg border border-slate-700 bg-slate-950 p-4">
      <h2 className="mb-4 text-lg font-medium">Donations</h2>

      {error ? (
        <p className="mb-4 rounded-md border border-rose-700 bg-rose-950/40 p-3 text-sm text-rose-200">
          {error}
        </p>
      ) : null}

      {actionError ? (
        <p className="mb-4 rounded-md border border-rose-700 bg-rose-950/40 p-3 text-sm text-rose-200">
          {actionError}
        </p>
      ) : null}

      {sortedDonations.length === 0 ? (
        <p className="rounded-md border border-slate-700 bg-slate-900/40 p-4 text-sm text-slate-300">
          No donations found yet.
        </p>
      ) : null}

      <div className="hidden md:block">
        <table className="w-full table-fixed border-collapse text-sm">
          <thead>
            <tr className="border-b border-slate-700 text-left text-slate-300">
              <th className="px-3 py-2 font-medium">
                <button
                  type="button"
                  onClick={() => onSort("uuid")}
                  className="text-left text-slate-300 hover:text-white"
                >
                  UUID ({sortLabel("uuid")})
                </button>
              </th>
              <th className="px-3 py-2 font-medium">Status</th>
              <th className="px-3 py-2 font-medium">
                <button
                  type="button"
                  onClick={() => onSort("amount")}
                  className="text-left text-slate-300 hover:text-white"
                >
                  Amount ({sortLabel("amount")})
                </button>
              </th>
              <th className="px-3 py-2 font-medium">Payment Method</th>
              <th className="px-3 py-2 font-medium">
                <button
                  type="button"
                  onClick={() => onSort("nonprofitId")}
                  className="text-left text-slate-300 hover:text-white"
                >
                  Nonprofit ({sortLabel("nonprofitId")})
                </button>
              </th>
              <th className="px-3 py-2 font-medium">
                <button
                  type="button"
                  onClick={() => onSort("createdAt")}
                  className="text-left text-slate-300 hover:text-white"
                >
                  Created ({sortLabel("createdAt")})
                </button>
              </th>
              <th className="px-3 py-2 font-medium">
                <button
                  type="button"
                  onClick={() => onSort("updatedAt")}
                  className="text-left text-slate-300 hover:text-white"
                >
                  Updated ({sortLabel("updatedAt")})
                </button>
              </th>
              <th className="px-3 py-2 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {paginatedDonations.map((donation) => {
              const availableActions = statusTransitions[donation.status];
              return (
                <tr key={donation.uuid} className="border-b border-slate-800 align-top">
                  <td className="break-all px-3 py-2 font-mono text-xs text-slate-200">
                    {donation.uuid}
                  </td>
                  <td className="px-3 py-2 text-slate-200">{donation.status}</td>
                  <td className="px-3 py-2 text-slate-100">{toDollars(donation.amount)}</td>
                  <td className="px-3 py-2 text-slate-200">{donation.paymentMethod}</td>
                  <td className="px-3 py-2 text-slate-200">{donation.nonprofitId}</td>
                  <td className="px-3 py-2 text-slate-200">
                    {toLocalDateTime(donation.createdAt)}
                  </td>
                  <td className="px-3 py-2 text-slate-200">
                    {toLocalDateTime(donation.updatedAt)}
                  </td>
                  <td className="px-3 py-2">
                    {availableActions.length === 0 ? (
                      <span className="text-xs text-slate-400">No actions</span>
                    ) : (
                      <div className="flex flex-wrap gap-2">
                        {availableActions.map((targetStatus) => (
                          <button
                            key={`${donation.uuid}-${targetStatus}`}
                            type="button"
                            disabled={pendingUUID === donation.uuid}
                            onClick={() => onStatusChange(donation, targetStatus)}
                            className="rounded-md border border-slate-600 bg-slate-900 px-2 py-1 text-xs text-slate-100 hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
                          >
                            Mark {targetStatus}
                          </button>
                        ))}
                      </div>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="space-y-3 md:hidden">
        {paginatedDonations.map((donation) => {
          const availableActions = statusTransitions[donation.status];
          return (
            <article
              key={donation.uuid}
              className="rounded-md border border-slate-700 bg-slate-900/50 p-3"
            >
              <p className="mb-2 break-all font-mono text-xs text-slate-300">{donation.uuid}</p>
              <div className="grid grid-cols-2 gap-y-2 text-sm">
                <p className="text-slate-400">Status</p>
                <p className="text-slate-200">{donation.status}</p>
                <p className="text-slate-400">Amount</p>
                <p className="text-slate-100">{toDollars(donation.amount)}</p>
                <p className="text-slate-400">Payment</p>
                <p className="text-slate-200">{donation.paymentMethod}</p>
                <p className="text-slate-400">Nonprofit</p>
                <p className="text-slate-200">{donation.nonprofitId}</p>
                <p className="text-slate-400">Created</p>
                <p className="text-slate-200">{toLocalDateTime(donation.createdAt)}</p>
                <p className="text-slate-400">Updated</p>
                <p className="text-slate-200">{toLocalDateTime(donation.updatedAt)}</p>
              </div>

              <div className="mt-3">
                {availableActions.length === 0 ? (
                  <span className="text-xs text-slate-400">No actions</span>
                ) : (
                  <div className="flex flex-wrap gap-2">
                    {availableActions.map((targetStatus) => (
                      <button
                        key={`${donation.uuid}-${targetStatus}`}
                        type="button"
                        disabled={pendingUUID === donation.uuid}
                        onClick={() => onStatusChange(donation, targetStatus)}
                        className="rounded-md border border-slate-600 bg-slate-900 px-2 py-1 text-xs text-slate-100 hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        Mark {targetStatus}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </article>
          );
        })}
      </div>

      {sortedDonations.length > 0 ? (
        <div className="mt-4 flex items-center justify-between gap-3 border-t border-slate-800 pt-3 text-sm text-slate-300">
          <p>
            Showing {paginationStart}-{paginationEnd} of {sortedDonations.length}
          </p>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}
              disabled={effectiveCurrentPage === 1}
              className="rounded-md border border-slate-600 bg-slate-900 px-3 py-1 text-xs text-slate-100 hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Previous
            </button>
            <span className="text-xs text-slate-300">
              Page {effectiveCurrentPage} of {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}
              disabled={effectiveCurrentPage === totalPages}
              className="rounded-md border border-slate-600 bg-slate-900 px-3 py-1 text-xs text-slate-100 hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Next
            </button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
