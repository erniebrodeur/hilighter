const copyButtons = document.querySelectorAll('[data-copy-target]');

copyButtons.forEach((button) => {
  button.addEventListener('click', async () => {
    const target = document.getElementById(button.dataset.copyTarget);
    if (!target) return;

    try {
      await navigator.clipboard.writeText(target.textContent.trim());
      const previousLabel = button.textContent;
      button.textContent = 'Copied';
      button.classList.add('copied');
      window.setTimeout(() => {
        button.textContent = previousLabel;
        button.classList.remove('copied');
      }, 1400);
    } catch {
      const selection = window.getSelection();
      const range = document.createRange();
      range.selectNodeContents(target);
      selection.removeAllRanges();
      selection.addRange(range);
    }
  });
});

document.getElementById('year').textContent = new Date().getFullYear();
