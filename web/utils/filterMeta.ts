export type FilterMeta = {
  title: string
  hint?: string
}

export const FILTER_META: Record<string, FilterMeta> = {
  remote_only: { title: 'Remote only (legacy)', hint: 'Shortcut: only fully remote listings' },
  arrangement: { title: 'Work arrangement', hint: 'OR within this group — remote, hybrid, or onsite' },
  weekly_days: { title: 'Weekly schedule', hint: 'e.g. 4-day or 4.5-day week' },
  async_friendly: { title: 'Async / flexible hours', hint: 'Listing signals async-first culture' },
  timezone_overlap_hours: { title: 'Timezone overlap', hint: 'Minimum overlap hours with your zone' },
  regions: { title: 'Regions', hint: 'Broad geography (SEA, EU, global, …)' },
  countries: { title: 'Countries', hint: 'Narrow to specific countries' },
  exclude_regions: { title: 'Exclude regions', hint: 'Hard-drop or de-prioritize regions' },
  employment_type: { title: 'Employment type', hint: 'Full-time, contract, etc.' },
  exclude_employment: { title: 'Exclude employment types', hint: 'e.g. commission-only, unpaid' },
  seniority: { title: 'Seniority', hint: 'Match level in title or metadata' },
  min_seniority: { title: 'Minimum seniority', hint: 'Floor for seniority rank' },
  salary_min: { title: 'Minimum salary', hint: 'Annual floor with currency' },
  salary_disclosed: { title: 'Salary disclosed', hint: 'Posting shows compensation' },
  equity: { title: 'Equity offered', hint: 'Role mentions equity' },
  company_stage: { title: 'Company stage', hint: 'Seed through public' },
  company_size: { title: 'Company size', hint: 'Headcount band' },
  visa_sponsorship: { title: 'Visa sponsorship', hint: 'Sponsorship mentioned' },
  relocation: { title: 'Relocation', hint: 'Relocation package offered' },
  eor_friendly: { title: 'EOR / contractor-friendly', hint: 'Employer-of-record or contractor OK' },
  archetype: { title: 'Role archetype', hint: 'Backend, fullstack, platform, …' },
  must_have_keywords: { title: 'Must-have keywords', hint: 'AND — all must appear' },
  nice_to_have_keywords: { title: 'Nice-to-have keywords', hint: 'OR — any match boosts score' },
  exclude_keywords: { title: 'Exclude keywords', hint: 'Drop if these appear' },
  benefits: { title: 'Benefits', hint: '4-day week, visa, equity tags' },
  max_posted_days: { title: 'Max age (days)', hint: 'How fresh the posting must be' },
}

export function filterTitle(key: string): string {
  return FILTER_META[key]?.title ?? key.replace(/_/g, ' ')
}
