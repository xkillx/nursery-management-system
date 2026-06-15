// Temporary child-detail mock fields used to render the full UI design while
// backend support is added. Replace each field with API-backed data before release.
export interface ChildDetailMockDocument {
  name: string;
  meta: string;
}

export interface ChildDetailMockProfile {
  roomName: string;
  childReference: string;
  attendanceSummary: string;
  alertChips: string[];
  documents: ChildDetailMockDocument[];
}

export const mockChildDetailProfile: ChildDetailMockProfile = {
  roomName: 'Toddlers Room',
  childReference: 'NP-7721-LA',
  attendanceSummary: 'Attendance opens in the practitioner register',
  alertChips: ['Peanut allergy - severe', 'Asthma inhaler in bag', 'Dietary: no gelatine'],
  documents: [
    { name: 'Registration Form.pdf', meta: 'Pending document storage' },
    { name: 'Consent Record.pdf', meta: 'Pending document storage' },
    { name: 'Medical History.pdf', meta: 'Pending document storage' },
  ],
};
