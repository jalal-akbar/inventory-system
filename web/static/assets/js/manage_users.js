function openAddModal() {
    document.getElementById('addModal').style.display = 'flex';
}

function openEditModal(user) {
    document.getElementById('edit_user_id').value = user.id;
    document.getElementById('edit_username').value = user.username;
    document.getElementById('edit_role').value = user.role;
    document.getElementById('edit_status').value = user.status;
    document.getElementById('editModal').style.display = 'flex';
}

function closeModal(id) {
    document.getElementById(id).style.display = 'none';
}

function confirmDelete(id, username) {
    Swal.fire({
        title: 'Are you sure?',
        text: `You are about to inactivate user: ${username}. They will no longer be able to login.`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#ef4444',
        cancelButtonColor: '#64748b',
        confirmButtonText: 'Yes, Inactivate'
    }).then((result) => {
        if (result.isConfirmed) {
            document.getElementById('formAction').value = 'delete';
            document.getElementById('formUserId').value = id;
            document.getElementById('actionForm').submit();
        }
    });
}

function confirmActivate(id, username) {
    Swal.fire({
        title: 'Activate User?',
        text: `Restore access for user: ${username}?`,
        icon: 'question',
        showCancelButton: true,
        confirmButtonColor: '#10b981',
        cancelButtonColor: '#64748b',
        confirmButtonText: 'Yes, Activate'
    }).then((result) => {
        if (result.isConfirmed) {
            document.getElementById('formAction').value = 'activate';
            document.getElementById('formUserId').value = id;
            document.getElementById('actionForm').submit();
        }
    });
}

window.onclick = function (event) {
    if (event.target.classList.contains('modal-overlay')) {
        event.target.style.display = 'none';
    }
}
