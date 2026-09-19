(() => {
  const container = document.querySelector('.filters');
  if (!container) return;

  const buttons = Array.from(container.querySelectorAll('.filter-btn'));
  const items = Array.from(document.querySelectorAll('.card, .writing-item'));
  const groups = Array.from(document.querySelectorAll('.project-group, .writing-section'));
  const noResults = document.querySelector('#no-results');

  if (buttons.length <= 1 || items.length === 0) return;

  container.hidden = false;

  const setFilter = (selectedTag) => {
    let totalVisible = 0;

    buttons.forEach((btn) => {
      const active = btn.dataset.tag === selectedTag;
      btn.classList.toggle('is-active', active);
      btn.setAttribute('aria-pressed', active ? 'true' : 'false');
    });

    groups.forEach((group) => {
      let groupVisible = 0;
      const groupItems = group.querySelectorAll('.card, .writing-item');

      groupItems.forEach((item) => {
        const rawTags = item.dataset.tags || '';
        const tags = rawTags.split(',').map((t) => t.trim());
        const match = !selectedTag || tags.includes(selectedTag);

        item.hidden = !match;
        if (match) groupVisible++;
      });

      group.hidden = groupVisible === 0;
      const countEl = group.querySelector('.group-count');
      if (countEl && groupVisible > 0) {
        countEl.textContent = groupVisible;
      }
      totalVisible += groupVisible;
    });

    if (noResults) {
      noResults.hidden = totalVisible > 0;
    }
  };

  container.addEventListener('click', (e) => {
    const btn = e.target.closest('.filter-btn');
    if (!btn) return;
    setFilter(btn.dataset.tag || '');
  });
})();

(() => {
  const toggle = document.querySelector('#theme-toggle');
  if (!toggle) return;
  toggle.addEventListener('click', () => {
    const current = document.documentElement.getAttribute('data-theme');
    const isDark = current === 'dark' || (!current && window.matchMedia('(prefers-color-scheme: dark)').matches);
    const next = isDark ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    try {
      localStorage.setItem('alster-theme', next);
    } catch (e) {}
  });
})();

