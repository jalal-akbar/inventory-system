/**
 * Password Toggle Functionality
 */

function initPasswordToggle() {
    const passwordInputs = document.querySelectorAll('input[type="password"][data-toggle-password]');

    passwordInputs.forEach(input => {
        // Buat wrapper untuk input dan tombol
        const wrapper = document.createElement('div');
        wrapper.className = 'password-input-wrapper';

        // Pindahkan input ke dalam wrapper
        input.parentNode.insertBefore(wrapper, input);
        wrapper.appendChild(input);

        // Buat tombol toggle
        const toggleBtn = document.createElement('button');
        toggleBtn.type = 'button';
        toggleBtn.className = 'password-toggle-btn';
        toggleBtn.innerHTML = `
            <svg class="eye-icon eye-off" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                <line x1="1" y1="1" x2="23" y2="23"></line>
            </svg>
            <svg class="eye-icon eye-on" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="display: none;">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                <circle cx="12" cy="12" r="3"></circle>
            </svg>
        `;

        wrapper.appendChild(toggleBtn);

        // Event listener untuk toggle
        toggleBtn.addEventListener('click', function (e) {
            e.preventDefault();

            const eyeOff = this.querySelector('.eye-off');
            const eyeOn = this.querySelector('.eye-on');

            if (input.type === 'password') {
                input.type = 'text';
                eyeOff.style.display = 'none';
                eyeOn.style.display = 'block';
            } else {
                input.type = 'password';
                eyeOff.style.display = 'block';
                eyeOn.style.display = 'none';
            }
        });
    });
}

// Auto-init ketika DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initPasswordToggle);
} else {
    initPasswordToggle();
}
