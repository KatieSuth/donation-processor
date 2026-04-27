import DonationsTable from "@/app/_components/DonationsTable";

export default async function Page() {
  return (
    <main className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-6">
      <div className="w-full max-w-2xl rounded-xl border border-slate-800 bg-slate-900 p-8">
        <h1 className="text-3xl font-semibold mb-2">Donations Dashboard</h1>
        <p className="text-slate-300 mb-6">
          View all donations and update statuses based on valid transitions.
        </p>
        <DonationsTable />
      </div>
    </main>
  );
}
