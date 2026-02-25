/**
 * Main JS - Apotek Desktop
 */

document.addEventListener('DOMContentLoaded', () => {
    updatePendingCount();
    const searchInput = document.getElementById('globalSearch');
    const searchResults = document.getElementById('searchResults');
    let debounceTimer;

    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            clearTimeout(debounceTimer);
            const query = e.target.value.trim();

            if (query.length < 2) {
                searchResults.style.display = 'none';
                return;
            }

            debounceTimer = setTimeout(() => {
                fetchDrugs(query);
            }, 300);
        });

        // Close search results when clicking outside
        document.addEventListener('click', (e) => {
            if (!searchInput.contains(e.target) && !searchResults.contains(e.target)) {
                searchResults.style.display = 'none';
            }
        });
    }

    async function fetchDrugs(query) {
        try {
            const response = await fetch(`api/drugs/search.php?q=${encodeURIComponent(query)}`);
            if (!response.ok) throw new Error('Network response was not ok');

            const data = await response.json();
            renderSearchResults(data);
        } catch (error) {
            console.error('Search error:', error);
        }
    }

    function renderSearchResults(results) {
        if (results.length === 0) {
            searchResults.innerHTML = '<div class="search-result-item">No drugs found.</div>';
        } else {
            searchResults.innerHTML = results.map(drug => `
                <div class="search-result-item" onclick="window.location.href='inventory.php?id=${drug.id}'">
                    <div class="result-header">
                        <span class="result-name">${drug.name}</span>
                        <span class="result-price">${drug.formatted_price}</span>
                    </div>
                    <div class="result-details">
                        <span><i data-lucide="tag" style="width:12px; height:12px; display:inline-block; vertical-align:middle; margin-right:4px;"></i>${drug.category}</span>
                        <span><i data-lucide="map-pin" style="width:12px; height:12px; display:inline-block; vertical-align:middle; margin-right:4px;"></i>${drug.location}</span>
                        <span><i data-lucide="archive" style="width:12px; height:12px; display:inline-block; vertical-align:middle; margin-right:4px;"></i>Stock: ${drug.stock}</span>
                    </div>
                </div>
            `).join('');

            // Re-init icons for dynamic content
            if (window.lucide) {
                lucide.createIcons();
            }
        }
        searchResults.style.display = 'block';
    }

    // Notification Dropdown Elements
    const notificationBell = document.getElementById('pendingBell');
    const notificationMenu = document.querySelector('.notification-menu');

    // User Profile Dropdown Logic
    const userProfile = document.getElementById('userProfileDropdown');
    const userDropdown = document.getElementById('userDropdownContent');

    if (userProfile && userDropdown) {
        userProfile.addEventListener('click', (e) => {
            e.stopPropagation();
            userDropdown.classList.toggle('show');
            userProfile.classList.toggle('active');

            // Close notification menu if open
            if (notificationMenu) {
                notificationMenu.classList.remove('show');
                if (notificationBell) notificationBell.classList.remove('active');
            }
        });
    }

    if (notificationBell && notificationMenu) {
        notificationBell.addEventListener('click', (e) => {
            e.stopPropagation();
            notificationMenu.classList.toggle('show');
            notificationBell.classList.toggle('active');

            // Close user dropdown if open
            if (userDropdown) {
                userDropdown.classList.remove('show');
                if (userProfile) userProfile.classList.remove('active');
            }
        });
    }

    // Close dropdowns when clicking outside
    document.addEventListener('click', (e) => {
        if (userProfile && userDropdown && !userProfile.contains(e.target)) {
            userDropdown.classList.remove('show');
            userProfile.classList.remove('active');
        }
        if (notificationMenu && notificationBell && !notificationBell.contains(e.target) && !notificationMenu.contains(e.target)) {
            notificationMenu.classList.remove('show');
            notificationBell.classList.remove('active');
        }
    });

    // Global Live Clock Logic
    const clockContainer = document.getElementById('clock-container');
    const clockTime = document.getElementById('global-live-clock');
    const clockDate = document.getElementById('global-live-date');

    if (clockContainer && clockTime && clockDate) {
        const timezone = clockContainer.getAttribute('data-timezone') || 'UTC';
        const lang = clockContainer.getAttribute('data-lang') || 'en';

        function updateGlobalClock() {
            const now = new Date();

            try {
                // Time Format (HH:mm:ss)
                const timeFormatter = new Intl.DateTimeFormat(lang, {
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit',
                    hour12: false,
                    timeZone: timezone
                });

                // Date Format
                const dateFormatter = new Intl.DateTimeFormat(lang, {
                    day: '2-digit',
                    month: 'short',
                    year: 'numeric',
                    timeZone: timezone
                });

                clockTime.textContent = timeFormatter.format(now);
                clockDate.textContent = dateFormatter.format(now);
            } catch (e) {
                console.error('Clock error:', e);
                // Fallback to basic string if Intl fails
                clockTime.textContent = now.toLocaleTimeString();
                clockDate.textContent = now.toLocaleDateString();
            }
        }

        updateGlobalClock();
        setInterval(updateGlobalClock, 1000);
    }

    // Sidebar Toggle Logic
    const sidebarToggle = document.getElementById('sidebarToggle');
    if (sidebarToggle) {
        // Load initial state from localStorage
        if (localStorage.getItem('sidebar-collapsed') === 'true') {
            document.body.classList.add('sidebar-collapsed');
        }

        sidebarToggle.addEventListener('click', () => {
            document.body.classList.toggle('sidebar-collapsed');
            localStorage.setItem('sidebar-collapsed', document.body.classList.contains('sidebar-collapsed'));
        });
    }

});

function updatePendingCount() {
    const badge = document.getElementById('pendingCount');
    if (!badge) return;

    fetch('api/drugs/pending-count')
        .then(res => res.json())
        .then(data => {
            badge.innerText = data.count;
            if (data.count > 0) {
                badge.style.display = 'flex';
                if (notificationBell) notificationBell.classList.add('has-notifications');
            } else {
                badge.style.display = 'none';
                if (notificationBell) notificationBell.classList.remove('has-notifications');
            }
        })
        .catch(err => console.error('Error fetching pending count:', err));
}
