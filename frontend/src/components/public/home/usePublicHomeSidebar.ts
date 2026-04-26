import { computed, ref } from "vue";
import { httpClient } from "../../../lib/http/client";
import type {
  AnnouncementItem,
  HotDownloadItem,
  LatestItem,
  PublicFileItem,
  SidebarDetailModalState,
} from "./types";
import type { InfoPanelCardItem } from "../../shared/InfoPanelCard.vue";

export function usePublicHomeSidebar(
  openFile: (fileID: string) => void,
  syncBodyScrollLock: () => void,
) {
  const announcements = ref<AnnouncementItem[]>([]);
  const announcementDetail = ref<AnnouncementItem | null>(null);
  const announcementListOpen = ref(false);
  const hotDownloadItems = ref<HotDownloadItem[]>([]);
  const latestItems = ref<LatestItem[]>([]);
  const sidebarDetailModal = ref<SidebarDetailModalState | null>(null);

  const hotDownloads = computed(() =>
    hotDownloadItems.value.slice(0, 5).map((item) => ({
      id: item.id,
      label: item.name,
    })),
  );

  const latestTitles = computed(() =>
    latestItems.value.slice(0, 5).map((item) => ({
      id: item.id,
      label: item.name,
    })),
  );

  const recentAnnouncements = computed(() =>
    announcements.value.slice(0, 5).map((item) => ({
      id: item.id,
      label: item.title,
      badge: item.is_pinned ? "置顶" : undefined,
    })),
  );

  async function loadAnnouncements() {
    try {
      const response = await httpClient.get<{ items: AnnouncementItem[] }>(
        "/public/announcements",
      );
      announcements.value = response.items ?? [];
    } catch {
      announcements.value = [];
    }
  }

  async function loadHotDownloads() {
    try {
      const response = await httpClient.get<{ items: PublicFileItem[] }>(
        "/public/files/hot?limit=20",
      );
      hotDownloadItems.value = (response.items ?? []).map((item) => ({
        id: item.id,
        name: item.name,
        downloadCount: item.download_count ?? 0,
      }));
    } catch {
      hotDownloadItems.value = [];
    }
  }

  async function loadLatestTitles() {
    try {
      const response = await httpClient.get<{ items: PublicFileItem[] }>(
        "/public/files/latest?limit=20",
      );
      latestItems.value = (response.items ?? []).map((item) => ({
        id: item.id,
        name: item.name,
      }));
    } catch {
      latestItems.value = [];
    }
  }

  function openSidebarDetailModal(modal: SidebarDetailModalState) {
    sidebarDetailModal.value = modal;
    syncBodyScrollLock();
  }

  function closeSidebarDetailModal() {
    sidebarDetailModal.value = null;
    syncBodyScrollLock();
  }

  function openSidebarDetailItem(item: InfoPanelCardItem) {
    sidebarDetailModal.value = null;
    syncBodyScrollLock();
    openFile(item.id);
  }

  function openAnnouncementDetail(item: InfoPanelCardItem) {
    const target = announcements.value.find((entry) => entry.id === item.id);
    if (!target) {
      return;
    }
    announcementListOpen.value = false;
    announcementDetail.value = target;
    syncBodyScrollLock();
  }

  function closeAnnouncementDetail() {
    announcementDetail.value = null;
    syncBodyScrollLock();
  }

  function returnToAnnouncementList() {
    announcementDetail.value = null;
    announcementListOpen.value = true;
    syncBodyScrollLock();
  }

  function openAnnouncementList() {
    announcementListOpen.value = true;
    syncBodyScrollLock();
  }

  function closeAnnouncementList() {
    announcementListOpen.value = false;
    syncBodyScrollLock();
  }

  function openHotDownloadsModal() {
    openSidebarDetailModal({
      eyebrow: "Hot Downloads",
      title: "热门下载",
      description: "展示近七天内下载量最高的前 20 份资料，点击可跳转文件详情页。",
      items: hotDownloadItems.value.map((item) => ({
        id: item.id,
        label: item.name,
        meta: `${item.downloadCount} 次下载`,
      })),
    });
  }

  function openLatestItemsModal() {
    openSidebarDetailModal({
      eyebrow: "Latest Files",
      title: "资料上新",
      description: "展示最新发布的前 20 份资料，点击标题可跳转文件详情页。",
      items: latestItems.value.map((item) => ({
        id: item.id,
        label: item.name,
      })),
    });
  }

  return {
    announcementDetail,
    announcementListOpen,
    announcements,
    closeAnnouncementDetail,
    closeAnnouncementList,
    closeSidebarDetailModal,
    hotDownloads,
    hotDownloadItems,
    latestItems,
    latestTitles,
    loadAnnouncements,
    loadHotDownloads,
    loadLatestTitles,
    openAnnouncementDetail,
    openAnnouncementList,
    openHotDownloadsModal,
    openLatestItemsModal,
    openSidebarDetailItem,
    recentAnnouncements,
    returnToAnnouncementList,
    sidebarDetailModal,
  };
}
