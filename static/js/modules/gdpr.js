import * as util from "./util.js";

const key = 'gdpr-consent';

export let consent = localStorage.getItem(key);

if (consent !== "yes" && consent !== "no") {
	const gdpr = document.querySelector("#gdpr_popup");
	util.hide(gdpr, false);

	const handler = e => {
		const choice = e.target.dataset["choice"];
		localStorage.setItem(key, choice)
		consent = choice;
		util.hide(gdpr, true);
	}
	const btns = gdpr.querySelectorAll("button[data-choice]");
	for (let i = 0; i < btns.length; i++) { btns[i].addEventListener("click", handler); }
}
