/**
 * Network Monitor
 *
 * Monitors network connectivity and notifies listeners of changes.
 * Detects when device goes online/offline.
 */

import NetInfo, { NetInfoState } from '@react-native-community/netinfo';
import { apiClient } from '../api';

export type NetworkStatus = 'online' | 'offline' | 'unknown';

export interface NetworkChangeListener {
  onNetworkChange: (status: NetworkStatus) => void;
}

class NetworkMonitorService {
  private listeners: Set<NetworkChangeListener> = new Set();
  private currentStatus: NetworkStatus = 'unknown';
  private unsubscribe: (() => void) | null = null;
  private lastStatusCheck: number = 0;
  private checkDebounceTime = 1000; // 1 second debounce

  /**
   * Start monitoring network changes
   */
  startMonitoring(): void {
    if (this.unsubscribe) {
      console.log('Network monitor already started');
      return;
    }

    console.log('Starting network monitor');

    // Subscribe to network state changes
    this.unsubscribe = NetInfo.addEventListener((state) => {
      this.handleNetworkStateChange(state);
    });

    // Check initial state
    this.checkNetworkState();
  }

  /**
   * Stop monitoring network changes
   */
  stopMonitoring(): void {
    if (this.unsubscribe) {
      this.unsubscribe();
      this.unsubscribe = null;
      console.log('Network monitor stopped');
    }
  }

  /**
   * Handle network state changes
   */
  private handleNetworkStateChange(state: NetInfoState): void {
    const now = Date.now();

    // Debounce rapid changes
    if (now - this.lastStatusCheck < this.checkDebounceTime) {
      return;
    }

    this.lastStatusCheck = now;

    // Determine new status
    const newStatus = this.determineStatus(state);

    // Only notify if status actually changed
    if (newStatus !== this.currentStatus) {
      console.log(`Network status changed: ${this.currentStatus} → ${newStatus}`);
      this.currentStatus = newStatus;
      this.notifyListeners(newStatus);
    }
  }

  /**
   * Determine network status from NetInfo state
   */
  private determineStatus(state: NetInfoState): NetworkStatus {
    if (state.isConnected === true && state.isInternetReachable !== false) {
      return 'online';
    } else if (state.isConnected === false) {
      return 'offline';
    } else {
      return 'unknown';
    }
  }

  /**
   * Check current network state
   */
  private async checkNetworkState(): Promise<void> {
    try {
      const state = await NetInfo.fetch();
      this.handleNetworkStateChange(state);
    } catch (error) {
      console.error('Failed to check network state:', error);
    }
  }

  /**
   * Add listener for network changes
   */
  addListener(listener: NetworkChangeListener): void {
    this.listeners.add(listener);
    console.log(`Network monitor listeners: ${this.listeners.size}`);
  }

  /**
   * Remove listener
   */
  removeListener(listener: NetworkChangeListener): void {
    this.listeners.delete(listener);
    console.log(`Network monitor listeners: ${this.listeners.size}`);
  }

  /**
   * Get current network status
   */
  getStatus(): NetworkStatus {
    return this.currentStatus;
  }

  /**
   * Check if online
   */
  isOnline(): boolean {
    return this.currentStatus === 'online';
  }

  /**
   * Check if offline
   */
  isOffline(): boolean {
    return this.currentStatus === 'offline';
  }

  /**
   * Force check network status
   */
  async forceCheck(): Promise<NetworkStatus> {
    await this.checkNetworkState();
    return this.currentStatus;
  }

  /**
   * Notify all listeners of status change
   */
  private notifyListeners(status: NetworkStatus): void {
    console.log(`Notifying ${this.listeners.size} listeners of status change: ${status}`);
    this.listeners.forEach((listener) => {
      try {
        listener.onNetworkChange(status);
      } catch (error) {
        console.error('Error notifying listener:', error);
      }
    });
  }

  /**
   * Verify connection by testing server
   */
  async verifyConnection(): Promise<boolean> {
    try {
      console.log('Verifying server connection...');
      const connected = await apiClient.testConnection();
      console.log(`Server connection verified: ${connected}`);
      return connected;
    } catch (error) {
      console.error('Connection verification failed:', error);
      return false;
    }
  }
}

// Export singleton instance
export const networkMonitor = new NetworkMonitorService();
