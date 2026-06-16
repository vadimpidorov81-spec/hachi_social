const card = document.querySelector(".status-card");
const label = document.querySelector("#status-label");
const detail = document.querySelector("#status-detail");
const profileForm = document.querySelector("#profile-form");
const formStatus = document.querySelector("#form-status");

const profileElements = {
  avatar: document.querySelector("#profile-avatar"),
  name: document.querySelector("#profile-name"),
  username: document.querySelector("#profile-username"),
  bio: document.querySelector("#profile-bio"),
  role: document.querySelector("#profile-role"),
  timezone: document.querySelector("#profile-timezone"),
  displayNameInput: document.querySelector("#display-name"),
  bioInput: document.querySelector("#bio"),
  timezoneInput: document.querySelector("#timezone"),
};

async function checkPlatform() {
  try {
    const response = await fetch("/ready", { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`ready returned ${response.status}`);
    }
    card.classList.add("online");
    label.textContent = "Платформа работает";
    detail.textContent = "API и PostgreSQL доступны в локальном окружении.";
  } catch {
    card.classList.add("offline");
    label.textContent = "Платформа недоступна";
    detail.textContent = "Проверь Docker Compose и состояние PostgreSQL.";
  }
}

function renderProfile(user) {
  profileElements.avatar.textContent = user.display_name.slice(0, 1).toUpperCase();
  profileElements.name.textContent = user.display_name;
  profileElements.username.textContent = `@${user.username}`;
  profileElements.bio.textContent = user.bio || "Пользователь пока ничего о себе не написал.";
  profileElements.role.textContent = user.role;
  profileElements.timezone.textContent = user.timezone;
  profileElements.displayNameInput.value = user.display_name;
  profileElements.bioInput.value = user.bio;
  profileElements.timezoneInput.value = user.timezone;
}

async function loadProfile() {
  const response = await fetch("/api/v1/users/me", { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`profile returned ${response.status}`);
  }
  const payload = await response.json();
  renderProfile(payload.data);
}

profileForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const submit = profileForm.querySelector("button[type=submit]");
  submit.disabled = true;
  formStatus.textContent = "Сохраняем...";

  try {
    const response = await fetch("/api/v1/users/me", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        display_name: profileElements.displayNameInput.value,
        bio: profileElements.bioInput.value,
        timezone: profileElements.timezoneInput.value,
      }),
    });
    if (!response.ok) {
      throw new Error(`update returned ${response.status}`);
    }
    const payload = await response.json();
    renderProfile(payload.data);
    formStatus.textContent = "Сохранено";
  } catch {
    formStatus.textContent = "Не удалось сохранить";
  } finally {
    submit.disabled = false;
  }
});

Promise.all([checkPlatform(), loadProfile()]).catch(() => {
  profileElements.name.textContent = "Профиль недоступен";
  profileElements.bio.textContent = "Проверь development identity и seed пользователя.";
});
