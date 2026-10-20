package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newShareCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Share AIM profiles across teammates over internal VPN",
		Long:  "Share AIM profiles across teammates over internal corporate VPN (Pritunl / OpenVPN).",
	}

	cmd.AddCommand(
		newShareEnableCmd(pm),
		newShareDisableCmd(pm),
		newShareStatusCmd(pm),
		newShareListCmd(pm),
		newShareRequestCmd(pm),
		newShareReleaseCmd(pm),
		newSharePeerCmd(pm),
		newShareServeCmd(pm),
	)
	return cmd
}

func newShareEnableCmd(pm *profile.ProfileManager) *cobra.Command {
	var agent string
	var autoApprove bool
	var ttlStr string

	cmd := &cobra.Command{
		Use:   "enable <profile>",
		Short: "Enable sharing for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profName := args[0]
			ttl, err := time.ParseDuration(ttlStr)
			if err != nil {
				ttl = 4 * time.Hour
			}
			store := share.NewStore(config.BaseDir())
			if err := store.EnableProfile(profName, agent, autoApprove, ttl); err != nil {
				return err
			}
			fmt.Printf("%s Profile %s is now shared (agent=%s, auto_approve=%v, ttl=%s)\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				lipgloss.NewStyle().Foreground(tui.TextBright).Render(profName),
				agent, autoApprove, ttl)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agent, "agent", "a", "claude", "Agent type (claude, codex, gemini, agy)")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Automatically approve lease requests")
	cmd.Flags().StringVar(&ttlStr, "ttl", "4h", "Maximum lease duration")
	return cmd
}

func newShareDisableCmd(pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <profile>",
		Short: "Disable sharing for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profName := args[0]
			store := share.NewStore(config.BaseDir())
			if err := store.DisableProfile(profName); err != nil {
				return err
			}
			fmt.Printf("%s Profile %s is no longer shared\n",
				lipgloss.NewStyle().Foreground(tui.StatusYellow).Render("✔"),
				profName)
			return nil
		},
	}
}

func newShareStatusCmd(pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sharing status and active VPN interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			vpn, _ := share.DetectVPN()
			if vpn != nil && vpn.Connected {
				fmt.Printf("%s VPN: %s (%s, broadcast: %s)\n",
					lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
					vpn.InterfaceName, vpn.IP, vpn.BroadcastIP)
			} else {
				fmt.Printf("%s VPN: Disconnected (Pritunl not active)\n",
					lipgloss.NewStyle().Foreground(tui.StatusYellow).Render("!"))
			}

			store := share.NewStore(config.BaseDir())
			profs := store.GetSharedProfiles()
			if len(profs) == 0 {
				fmt.Println("No profiles currently shared.")
				return nil
			}
			fmt.Println(presenter.RenderSharedProfilesTable(os.Stdout, profs))
			return nil
		},
	}
}

func newShareListCmd(pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Discover peers and available shared profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := share.NewStore(config.BaseDir())
			peers := store.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers discovered yet. Use 'aim share peer add <vpn-ip>' or ensure teammates are on the VPN.")
				return nil
			}
			fmt.Println(presenter.RenderPeersTable(os.Stdout, peers))
			return nil
		},
	}
}

func newShareRequestCmd(pm *profile.ProfileManager) *cobra.Command {
	var hours int
	var agent string

	cmd := &cobra.Command{
		Use:   "request <peer> <profile> (or <peer>:<profile> / <peer>/<profile>)",
		Short: "Request a lease for a peer's profile",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var peer, profileName string
			if len(args) == 2 {
				peer = args[0]
				profileName = args[1]
			} else if strings.Contains(args[0], "/") {
				parts := strings.SplitN(args[0], "/", 2)
				peer, profileName = parts[0], parts[1]
			} else if strings.Contains(args[0], ":") {
				parts := strings.SplitN(args[0], ":", 2)
				peer, profileName = parts[0], parts[1]
			} else {
				return fmt.Errorf("format must be <peer> <profile>, <peer>/<profile>, or <peer>:<profile>")
			}

			store := share.NewStore(config.BaseDir())
			targetHost := peer
			for _, p := range store.GetPeers() {
				shortHost := strings.TrimSuffix(strings.ToLower(p.HostName), ".local")
				if strings.EqualFold(p.HostName, peer) || strings.EqualFold(shortHost, strings.ToLower(peer)) || p.VPNIP == peer {
					if p.HostName != "" {
						targetHost = p.HostName
					} else {
						targetHost = p.VPNIP
					}
					break
				}
			}

			client := share.NewClient(4319)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := client.RequestLease(ctx, targetHost, profileName, agent, hours)
			if err != nil && !strings.Contains(targetHost, ".") {
				// Retry with .local for zero-conf mDNS resolution
				resp, err = client.RequestLease(ctx, targetHost+".local", profileName, agent, hours)
			}
			if err != nil {
				return fmt.Errorf("lease request failed: %w", err)
			}
			if resp.Status != "approved" {
				fmt.Printf("%s Request status: %s (%s)\n",
					lipgloss.NewStyle().Foreground(tui.StatusYellow).Render("!"),
					resp.Status, resp.Message)
				return nil
			}

			mgr := share.NewLeaseManager(pm.ProfilesRoot())
			localName, err := mgr.InstallLease(peer, profileName, resp)
			if err != nil {
				return fmt.Errorf("failed to install credentials: %w", err)
			}

			fmt.Printf("%s Lease granted until %s\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				resp.ExpiresAt.Format(time.RFC1123))
			fmt.Printf("Run natively with: %s\n",
				lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(fmt.Sprintf("aim run %s %s", agent, localName)))
			return nil
		},
	}
	cmd.Flags().IntVar(&hours, "hours", 2, "Requested lease duration in hours")
	cmd.Flags().StringVarP(&agent, "agent", "a", "claude", "Agent type")
	return cmd
}

func newShareReleaseCmd(pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "release <ephemeral-profile>",
		Short: "Release and remove a leased ephemeral profile early",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := share.NewLeaseManager(pm.ProfilesRoot())
			if err := mgr.ReleaseLease(args[0]); err != nil {
				return err
			}
			fmt.Printf("%s Leased profile %s released\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				args[0])
			return nil
		},
	}
}

func newSharePeerCmd(pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "Manage network peers",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "add <ip-or-host>",
		Short: "Add a teammate's VPN IP or hostname",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ip := args[0]
			store := share.NewStore(config.BaseDir())
			disc := share.NewDiscovery(store, 4319)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			peer, err := disc.PingPeer(ctx, ip)
			if err != nil {
				return fmt.Errorf("could not reach peer at %s: %w", ip, err)
			}
			fmt.Printf("%s Peer verified: %s (%s, profiles: %v)\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				peer.HostName, peer.VPNIP, peer.Profiles)
			return nil
		},
	})
	return cmd
}

func newShareServeCmd(pm *profile.ProfileManager) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the P2P sharing server listener",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := share.NewStore(config.BaseDir())
			host, _ := os.Hostname()
			server := share.NewServer(store, pm.ProfilesRoot(), host, port)

			addr := fmt.Sprintf("0.0.0.0:%d", port)
			fmt.Printf("%s AIM P2P Share Server listening on %s (press Ctrl+C to stop)\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				addr)

			go func() {
				if err := server.Start(addr); err != nil {
					fmt.Fprintf(os.Stderr, "server stopped: %v\n", err)
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			<-sigCh
			_ = server.Close()
			fmt.Println("\nShare server stopped.")
			return nil
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 4319, "Port to listen on")
	return cmd
}
