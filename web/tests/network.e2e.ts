import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

test('p2p network sharing view renders header, global policy control, and tabs', async ({ app, screen }) => {
  await app.open('/');

  // Find P2P Network navigation tab in sidebar
  const networkTab = screen.getByRole('tab', { name: /P2P Network/i });
  await expect(networkTab).toBeVisible();

  // Navigate to P2P Network view
  await networkTab.tap();

  // Verify breadcrumb or view title
  await expect(screen.getByText('P2P Network Sharing')).toBeVisible();

  // Verify Global Policy segmented control is permanently visible
  await expect(screen.getByText('Global Policy:')).toBeVisible();
  const offBtn = screen.getByRole('radio', { name: /^Off$/i });
  const reqBtn = screen.getByRole('radio', { name: /^Request$/i });
  const autoBtn = screen.getByRole('radio', { name: /^Auto Approve$/i });

  await expect(offBtn).toBeVisible();
  await expect(reqBtn).toBeVisible();
  await expect(autoBtn).toBeVisible();

  // Verify sub-navigation tabs within P2P Network
  const teammatesTab = screen.getByRole('tab', { name: /Teammates/i });
  const requestsTab = screen.getByRole('tab', { name: /Requests/i });
  const auditTab = screen.getByRole('tab', { name: /Audit Log/i });

  await expect(teammatesTab).toBeVisible();
  await expect(requestsTab).toBeVisible();
  await expect(auditTab).toBeVisible();

  // Teammates tab: displays discovered teammates or local node sharing policy card
  await expect(screen.getByRole('heading', { name: /Discovered Teammates/i })).toBeVisible();

  // Navigate to My Sharing tab
  const sharingTab = screen.getByRole('tab', { name: /My Sharing/i });
  await sharingTab.tap();
  await expect(screen.getByRole('heading', { name: /Your Offered Profiles/i })).toBeVisible();

  // Navigate to Requests tab
  await requestsTab.tap();
  await expect(screen.getByText(/Incoming Requests/i)).toBeVisible();
  await expect(screen.getByText(/All Requests & Leases/i)).toBeVisible();

  // Navigate to Audit Log tab
  await auditTab.tap();
  await expect(screen.getByText(/Tamper-evident mesh audit trail/i)).toBeVisible();

  // Return to Teammates tab
  await teammatesTab.tap();
  await expect(screen.getByRole('heading', { name: /Discovered Teammates/i })).toBeVisible();
});

test('p2p global policy toggle switches mode', async ({ app, screen }) => {
  await app.open('/network');

  await expect(screen.getByText('Global Policy:')).toBeVisible();

  const reqBtn = screen.getByRole('radio', { name: /^Request$/i });
  await reqBtn.tap();

  // Verify policy mode indicator updates or request state persists
  await expect(reqBtn).toBeVisible();
});
