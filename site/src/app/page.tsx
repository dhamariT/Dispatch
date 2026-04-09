"use client";

import { useMemo, useState } from "react";
import { DemoBanner } from "@/components/demo-banner";
import { DeployTabs } from "@/components/deploy-tabs";
import { DeployTopbar } from "@/components/deploy-topbar";
import { DeviceGroup } from "@/components/device-group";
import { DeviceListItem } from "@/components/device-list-item";
import { Nav } from "@/components/nav";
import { countFleetBySeverity, demoFleet } from "@/lib/demo-data";
import { getWorstSeverity } from "@/lib/severity";

export default function DashboardPage() {
  const [selectedDeployId, setSelectedDeployId] = useState(
    demoFleet.deploys[0].id,
  );
  const [expandedDevice, setExpandedDevice] = useState<string | null>("car-2");

  const selectedDeploy = useMemo(
    () =>
      demoFleet.deploys.find((d) => d.id === selectedDeployId) ??
      demoFleet.deploys[0],
    [selectedDeployId],
  );

  const counts = useMemo(
    () => countFleetBySeverity(selectedDeploy.devices),
    [selectedDeploy],
  );

  const sortedDevices = useMemo(() => {
    const rank = (
      device: (typeof selectedDeploy.devices)[number],
    ): number => {
      if (device.status === "offline") return 5;
      if (device.status === "waiting") return 4;
      const severity = getWorstSeverity(device.metrics);
      if (severity === "critical") return 0;
      if (severity === "warning") return 1;
      if (severity === "stable") return 2;
      return 3;
    };
    return [...selectedDeploy.devices].sort((a, b) => rank(a) - rank(b));
  }, [selectedDeploy]);

  const deployTabs = demoFleet.deploys.map((d) => ({
    id: d.id,
    from: d.from,
    to: d.to,
    severity: d.severity,
  }));

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Nav fleetName={demoFleet.name} />
      <DemoBanner />
      <DeployTopbar
        from={selectedDeploy.from}
        to={selectedDeploy.to}
        status={selectedDeploy.status}
        soakElapsed={selectedDeploy.soakElapsed}
        soakTotal={selectedDeploy.soakTotal}
        critical={counts.critical}
        warning={counts.warning}
        waiting={counts.waiting}
        offline={counts.offline}
        total={counts.total}
      />

      <main className="mx-auto flex w-full max-w-[1200px] flex-1 flex-col gap-6 px-8 py-8">
        <section className="flex flex-col gap-3">
          <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
            Deploy history
          </h2>
          <DeployTabs
            deploys={deployTabs}
            selected={selectedDeployId}
            onSelect={setSelectedDeployId}
          />
        </section>

        <section className="flex flex-col gap-3">
          <div className="flex items-baseline justify-between">
            <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
              Devices
            </h2>
            <span className="text-xs text-muted-foreground tabular-nums">
              Sorted by severity
            </span>
          </div>

          <div className="flex flex-col gap-2">
            {sortedDevices.map((device) => {
              const isExpanded = expandedDevice === device.deviceId;
              return (
                <div key={device.deviceId} className="flex flex-col gap-2">
                  <DeviceListItem
                    deviceId={device.deviceId}
                    status={device.status}
                    metrics={device.metrics}
                    expanded={isExpanded}
                    onClick={() =>
                      setExpandedDevice(isExpanded ? null : device.deviceId)
                    }
                  />
                  {isExpanded && device.metrics && (
                    <div className="pl-6 pr-2 pb-2">
                      <DeviceGroup
                        deviceId={device.deviceId}
                        status={device.status}
                        metrics={device.metrics}
                        tableOnly
                      />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </section>
      </main>
    </div>
  );
}
