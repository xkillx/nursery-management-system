export interface SessionTemplateEntry {
  id: string;
  dayOfWeek: number;
  sessionType: {
    id: string;
    name: string;
    startTime: string;
    endTime: string;
    isActive: boolean;
  };
}

export interface SessionTemplate {
  id: string;
  branchId: string;
  name: string;
  description?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  entries: SessionTemplateEntry[];
}

export interface SessionTemplateListItem {
  id: string;
  branchId: string;
  name: string;
  description?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  // List responses do not hydrate entries; consumers can fetch a single
  // template to inspect them.
  entries: never[];
}

export interface SessionTemplateInput {
  name: string;
  description?: string | null;
  entries: { dayOfWeek: number; sessionTypeId: string }[];
}

export interface SessionTemplateUpdateInput {
  name?: string;
  description?: string | null;
  entries?: { dayOfWeek: number; sessionTypeId: string }[];
}
