
import os
import time
import datetime  # To use the current time in printing
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException, NoSuchElementException
from xvfbwrapper import Xvfb
import traceback

# --- Settings ---
GMAIL_EMAIL = "you gmail" # use your gmail 
GMAIL_PASSWORD = "password gmail" #Gmail Password

BASE_DIR = "/root/YukkiMusic" #don't change it!
COOKIES_DIR = os.path.join(BASE_DIR, "cookies")
COOKIES_FILENAME = "cookies.txt"
COOKIES_PATH = os.path.join(COOKIES_DIR, COOKIES_FILENAME)

PROFILE_DIR_NAME = "chrome_profile"
PROFILE_PATH = os.path.join(BASE_DIR, PROFILE_DIR_NAME)

WAIT_TIMEOUT = 25
LOGIN_CHECK_TIMEOUT = 10

REFRESH_INTERVAL_HOURS = 2
INTERVAL_SECONDS = REFRESH_INTERVAL_HOURS * 60 * 60

def refresh_youtube_cookies_with_profile():
    """
    Uses the saved Chrome profile and updates (overwrites) the YouTube cookies file in the correct format without deleting it first.
    """
    current_time_str = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    print(f"\n{'='*15} Starting cookie refresh cycle ({current_time_str}) {'='*15}")
    print(f"-> Profile path: {PROFILE_PATH}")
    print(f"-> Cookies file path: {COOKIES_PATH}")

    # --- Ensure cookies directory exists ---
    try:
        os.makedirs(COOKIES_DIR, exist_ok=True)
    except OSError as e:
        print(f"!!! Critical error: Unable to create cookies directory {COOKIES_DIR}: {e}")
        return None

    # --- Setup Xvfb ---
    vdisplay = None
    try:
        vdisplay = Xvfb()
        vdisplay.start()
    except Exception as e:
        print(f"Warning: Xvfb is unavailable or failed to start. Error: {e}")

    # --- Setup Selenium WebDriver ---
    driver = None
    success = False
    try:
        print("-> Setting up Chrome browser options...")
        options = webdriver.ChromeOptions()
        options.add_argument("--headless=new")
        options.add_argument("--no-sandbox")
        options.add_argument("--disable-dev-shm-usage")
        options.add_argument("--window-size=1920x1080")
        options.add_argument('--disable-blink-features=AutomationControlled')
        options.add_experimental_option('excludeSwitches', ['enable-automation'])
        options.add_experimental_option('useAutomationExtension', False)
        options.add_argument("user-agent=Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
        options.add_argument(f"--user-data-dir={PROFILE_PATH}")
        options.add_argument("--profile-directory=Default")

        print("-> Installing/updating ChromeDriver...")
        service = Service(ChromeDriverManager().install())

        print("-> Starting Chrome browser with profile...")
        driver = webdriver.Chrome(service=service, options=options)
        driver.execute_script("Object.defineProperty(navigator, 'webdriver', {get: () => undefined})")
        driver.implicitly_wait(5)
        wait = WebDriverWait(driver, WAIT_TIMEOUT)
        print("-> Browser started.")

        # --- Check login status and navigate to YouTube ---
        print("-> Navigating to YouTube to check login status...")
        driver.get("https://www.youtube.com")
        time.sleep(3)

        logged_in = False
        try:
            signin_button_xpath = "//yt-button-renderer/a[contains(@href, 'ServiceLogin')] | //a[@aria-label='Sign in'] | //button[@aria-label='Sign in'] | //tp-yt-paper-button[contains(@aria-label,'Sign in')]"
            WebDriverWait(driver, LOGIN_CHECK_TIMEOUT).until(
                EC.presence_of_element_located((By.XPATH, signin_button_xpath))
            )
            print("   - 'Sign in' button is present. Login is required.")
            logged_in = False
        except TimeoutException:
            print("   - 'Sign in' button not found. Assuming session is active.")
            logged_in = True
        except Exception as check_err:
             print(f"   - Unexpected error while checking 'Sign in' button: {check_err}")
             print("   - Login will be attempted as a precaution.")
             logged_in = False

        # --- Login (only if necessary) ---
        if not logged_in:
            print("-> Starting Google login process...")
            driver.get("https://accounts.google.com/ServiceLogin?hl=en&passive=true&continue=https://www.youtube.com/signin?app=desktop&action_handle_signin=true&hl=en&next=%2F")
            time.sleep(2)
            try:
                print("   - Entering email...")
                email_field = wait.until(EC.visibility_of_element_located((By.XPATH, "//input[@type='email']")))
                email_field.send_keys(GMAIL_EMAIL)
                wait.until(EC.element_to_be_clickable((By.XPATH, "//*[text()='Next']/ancestor::button | //button[contains(.,'Next')] | //*[@id='identifierNext']"))).click()

                print("   - Entering password...")
                password_field = wait.until(EC.visibility_of_element_located((By.XPATH, "//input[@type='password']")))
                password_field.send_keys(GMAIL_PASSWORD)
                wait.until(EC.element_to_be_clickable((By.XPATH, "//*[text()='Next']/ancestor::button | //button[contains(.,'Next')] | //*[@id='passwordNext']"))).click()

                print("   - Waiting for login confirmation...")
                WebDriverWait(driver, WAIT_TIMEOUT).until(
                    EC.any_of(
                        EC.url_contains("youtube.com"),
                        EC.url_contains("myaccount.google.com"),
                        EC.staleness_of(password_field)
                    )
                )
                print("   - Successfully logged in (or step bypassed).")
                if "youtube.com" not in driver.current_url:
                    print("-> Not redirected to YouTube automatically, navigating now...")
                    driver.get("https://www.youtube.com")
                print("   - Waiting for page to load (5 seconds)...")
                time.sleep(5)

            except (TimeoutException, NoSuchElementException) as login_timeout_err:
                print(f"\n!!! Error: Failed to find field or button during login, or timeout occurred: {login_timeout_err} !!!")
                raise login_timeout_err
            except Exception as login_err:
                print(f"\n!!! Error during login process: {login_err} !!!")
                timestamp = time.strftime('%Y%m%d_%H%M%S')
                driver.save_screenshot(f"login_error_{timestamp}.png")
                with open(f"login_error_{timestamp}.html", "w", encoding='utf-8') as f_err:
                    f_err.write(driver.page_source)
                print(f"   - Saved screenshot and page source for the error.")
                raise login_err

        # --- Extract cookies ---
        print("-> Extracting updated cookies from YouTube...")
        cookies = driver.get_cookies()

        if not cookies:
            print("!!! Warning: No cookies found. Final attempt after waiting...")
            time.sleep(7)
            cookies = driver.get_cookies()
            if not cookies:
                 print("!!! Critical error: Failed to extract cookies completely.")
                 return None

        # --- *** Change here: No file deletion *** ---
        # Code to delete the old file was removed

        # --- Write new cookies (using 'w' mode will overwrite) ---
        print(f"-> Opening file {COOKIES_PATH} for writing (content will be replaced)...")
        written_count = 0
        try:
            # Using 'w' mode ensures old content is cleared before writing
            with open(COOKIES_PATH, "w", encoding='utf-8') as f:
                f.write("# Netscape HTTP Cookie File\n")
                f.write("# Generated by Selenium script for YouTube using profile persistence\n")
                f.write(f"# Profile: {PROFILE_PATH}\n")
                f.write(f"# Timestamp: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")

                for cookie in cookies:
                    if not all(k in cookie for k in ['name', 'value', 'domain', 'path']):
                        # print(f"   - Ignoring incomplete cookie: {cookie.get('name', 'N/A')}") # Uncomment for detailed logging
                        continue

                    domain = cookie.get('domain', '')
                    include_subdomains = str(domain.startswith('.')).upper()
                    path = cookie.get('path', '/')
                    secure = str(cookie.get('secure', False)).upper()
                    expiry_val = cookie.get('expiry')
                    expiry = int(expiry_val) if expiry_val is not None else 0
                    name = cookie.get('name', '')
                    value = str(cookie.get('value', ''))

                    if '\n' in name or '\n' in value or '\t' in name or '\t' in value:
                        # print(f"   - Ignoring cookie with invalid characters: Name='{name}'") # Uncomment for detailed logging
                        continue

                    f.write(f"{domain}\t{include_subdomains}\t{path}\t{secure}\t{expiry}\t{name}\t{value}\n")
                    written_count += 1

            print(f"   - Successfully wrote {written_count} cookies to {COOKIES_PATH}.")
            success = True
            return COOKIES_PATH

        except IOError as e:
            print(f"!!! Critical error while writing cookies file: {e} !!!")
            return None
        except Exception as format_err:
             print(f"!!! Unexpected error while formatting or writing cookie: {format_err} !!!")
             traceback.print_exc()
             return None

    except KeyboardInterrupt:
        print("\n[!] Received stop request (Ctrl+C) during cycle.")
        raise
    except Exception as e:
        print("\n" + "="*20 + " General unhandled error occurred in this cycle " + "="*20)
        print(f"Error type: {type(e).__name__}")
        print(f"Error message: {e}")
        traceback.print_exc()
        if driver:
            try:
                timestamp = time.strftime('%Y%m%d_%H%M%S')
                screenshot_path = os.path.join(BASE_DIR, f"general_error_{timestamp}.png")
                html_path = os.path.join(BASE_DIR, f"general_error_{timestamp}.html")
                driver.save_screenshot(screenshot_path)
                with open(html_path, "w", encoding='utf-8') as f_err:
                    f_err.write(driver.page_source)
                print(f"-> Saved screenshot ({screenshot_path}) and page source ({html_path}) for the general error.")
            except Exception as save_err:
                print(f"-> Failed to save additional error information: {save_err}")
        return None

    finally:
        # --- Cleanup ---
        if driver:
            try:
                driver.quit()
            except Exception as quit_err:
                 print(f"   - Warning: Error while closing browser: {quit_err}")
        if vdisplay:
            try:
                vdisplay.stop()
            except Exception as xvfb_stop_err:
                 print(f"   - Warning: Error while stopping Xvfb: {xvfb_stop_err}")
        print(f"{'='*15} End of refresh cycle {'='*15}")


# --- Main entry point and infinite loop ---
if __name__ == "__main__":
    print(f"[*] Starting YouTube cookie auto-refresher.")
    print(f"[*] Refresh interval: {REFRESH_INTERVAL_HOURS} hours ({INTERVAL_SECONDS} seconds).")
    print(f"[*] Profile path: {PROFILE_PATH}")
    print(f"[*] Cookies file path: {COOKIES_PATH}")
    print(f"[*] Write mode: Overwrite.")  # Clarify the method
    print("[*] Press Ctrl+C to stop.")

    while True:
        try:
            result = refresh_youtube_cookies_with_profile()

            if result:
                print(f"\n[Success] Refresh cycle completed successfully. Cookies updated at: {result}")
            else:
                print("\n[Failure] Refresh cycle did not complete successfully this time.")
                print("   - Will retry after the interval.")

            next_run_time = datetime.datetime.now() + datetime.timedelta(seconds=INTERVAL_SECONDS)
            print(f"\n[*] Waiting for {REFRESH_INTERVAL_HOURS} hours...")
            print(f"[*] Next cycle will start approximately at: {next_run_time.strftime('%Y-%m-%d %H:%M:%S')}")
            time.sleep(INTERVAL_SECONDS)

        except KeyboardInterrupt:
            print("\n[*] Stop request received (Ctrl+C). Exiting...")
            break

        except Exception as loop_error:
            print(f"\n!!! Critical error in main loop: {loop_error} !!!")
            traceback.print_exc()
            print("[!] Unexpected error occurred, will attempt to continue after one minute...")
            time.sleep(60)

    print("[*] YouTube cookie auto-refresher stopped.")
