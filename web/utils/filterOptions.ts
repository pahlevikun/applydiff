/** Known filter value vocabularies (OR within each category). */
export const FILTER_CATEGORY_OPTIONS: Record<string, string[]> = {
  arrangement: ['remote', 'hybrid', 'onsite'],
  regions: ['SEA', 'EU', 'APAC', 'EMEA', 'NA', 'LATAM', 'global', 'UK', 'US', 'AU', 'IN'],
  countries: ['ID', 'SG', 'MY', 'TH', 'VN', 'PH', 'UK', 'US', 'CA', 'AU', 'NZ', 'DE', 'NL'],
  exclude_regions: ['US', 'NA', 'LATAM', 'IN', 'MENA'],
  employment_type: ['full-time', 'part-time', 'contract', 'freelance', 'internship', 'contract-to-hire'],
  exclude_employment: ['commission-only', 'unpaid'],
  seniority: ['intern', 'junior', 'mid', 'senior', 'staff', 'principal', 'lead', 'manager', 'director'],
  min_seniority: ['intern', 'junior', 'mid', 'senior', 'staff', 'principal', 'lead'],
  company_stage: ['seed', 'series-a', 'series-b', 'series-c', 'public', 'bootstrapped'],
  company_size: ['1-10', '11-50', '51-200', '201-1000', '1000+'],
  archetype: ['backend', 'frontend', 'fullstack', 'mobile', 'platform', 'data', 'ml', 'devops', 'security', 'qa'],
  weekly_days: ['4', '4.5', '5'],
  benefits: ['4-day week', 'visa sponsorship', 'equity', 'remote-first', 'async-friendly'],
}

export function optionsForFilter(key: string, selected: string[] = []): string[] {
  const preset = FILTER_CATEGORY_OPTIONS[key] ?? []
  const seen = new Set(preset)
  const out = [...preset]
  for (const v of selected) {
    if (!seen.has(v)) {
      seen.add(v)
      out.push(v)
    }
  }
  return out
}
