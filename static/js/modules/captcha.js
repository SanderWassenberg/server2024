import { new_template, pick_random, random_in_range, random_int_in_range } from "./util.js";

const make_body = new_template(`
<canvas width="300" height="100"></canvas>
<form method="dialog">
<input type="text" placeholder="Copy text above">
</form>
<link rel="stylesheet" href="/css/captcha.css">
`)


export class Captcha extends HTMLElement {
	static { customElements.define("my-captcha", Captcha); }

	// elems
	#input;
	#ctx;

	// internal data
	#string;
	#failure_timeout; // handle of timeout which will enable the captcha again some time after a failed attempt
	#enable_func;     // function called by the timeout

	// public interface
	onsubmit;

	constructor() {
		super();

		this.attachShadow({mode: 'open'});
		this.shadowRoot.append(make_body());

		this.#input = this.shadowRoot.querySelector("input[type='text']");
		this.#ctx   = this.shadowRoot.querySelector("canvas").getContext("2d", {alpha: false});


		// this.shadowRoot.querySelector("#reset").addEventListener("click", () => {
		//	 this.reset();
		// })

		this.#input.addEventListener("keypress", e => {
			if (e.code !== "Enter") return;
			this.onsubmit?.();
		})

		this.#enable_func = () => {
			this.reset();
			this.#failure_timeout = undefined;
		}

		this.reset();
	}

	reset() {
		this.#input.value   = "";

		const str_len = 5;
		this.#input.maxLength = str_len;

		{
			// specific set of letters because some of the upper/lower case versions look too much alike
			const characters = "AaBbCDdEeFfGgHhIiJjKkLMmNnPpQqRrSTtUuVXYyZ123456789@#$%&";
			const arr = new Array(str_len);
			for (let i = 0; i < str_len; i++) {
				arr[i] = pick_random(characters);
			}
			this.#string = arr.join("");
		}

		const canvas = this.#ctx.canvas;

		const maxwidth = canvas.width / str_len
		const minwidth = 14
		const minheight = 25


		//this.#ctx.clearRect(0, 0, canvas.width, canvas.height);
		this.#ctx.beginPath();
		this.#ctx.fillStyle = "#aaa";
		this.#ctx.fillRect(0, 0, canvas.width, canvas.height);

		this.#ctx.lineWidth = 2;
		for (let i = 0; i < 100; i++) {
			this.#ctx.strokeStyle = `rgba(${random_int_in_range(0,200)},${random_int_in_range(0,200)},${random_int_in_range(0,200)},1)`;
			if (Math.random() < 0.66) {
				draw_line(this.#ctx, Math.random() * canvas.width, 0, Math.random() * canvas.width, canvas.height);
			} else {
				draw_line(this.#ctx, 0, Math.random() * canvas.height, canvas.width, Math.random() * canvas.height)
			}
		}

		for (let i = 0; i < str_len; i++) {
			this.#ctx.fillStyle = `rgba(0,0,0,0.6)`;
			const char_height = random_in_range(minheight, 0.8 * canvas.height);
			const char_width  = random_in_range(minwidth,  maxwidth);

			// times 0.9 to make sure letters that extend beyond the edge, like j and g are fully visible
			const y_pos = random_in_range(char_height, 0.9 * canvas.height)
			const x_pos = i * maxwidth;

			this.#ctx.font = `${char_height}px serif`;
			this.#ctx.fillText(this.#string[i], x_pos, y_pos, char_width)
		}
	}

	focus() {
		this.#input.focus()
	}

	#disable_temporarily() {
		clearTimeout(this.#failure_timeout);
		this.#ctx.clearRect(0, 0, this.#ctx.canvas.width, this.#ctx.canvas.height);
		this.#failure_timeout = setTimeout(this.#enable_func, 1000)
	}

	verify(err_handler) {
		if (this.#failure_timeout) {
			this.#disable_temporarily();
			err_handler?.("You're retrying the captcha too fast, please wait.");
			return false;
		}
		if (this.#input.value === "") {
			err_handler?.("Fill in the captcha.");
			return false;
		}

		const correct = this.#input.value === this.#string;

		if (!correct) {
			this.#disable_temporarily();
			err_handler?.("Incorrect. Try again.");
			return false;
		}

		this.reset();
		return true;
	}
}

function draw_line(ctx, x1, y1, x2, y2) {
	ctx.beginPath();
	ctx.moveTo(x1, y1);
	ctx.lineTo(x2, y2);
	ctx.stroke();
}