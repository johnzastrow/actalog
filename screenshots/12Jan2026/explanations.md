### image_001.jpg

<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_001.jpg" style="zoom: 25%;" />

We need to add the  RPE attribute to the Workout Templates and user workouts (for example as logged from the QuickLog screen, but also when logging WODs and Movements as user workouts). The RPE (Rate of Perceived Exertion) scale is a subjective tool (usually 0-10 or Borg 6-20) that measures how hard your body feels it's working during exercise, considering heart rate, breathing, and muscle fatigue, helping you gauge intensity without relying solely on gadgets, perfect for tailoring workouts to your fitness level, from light activity (2-3) to maximum effort (9-10). We will record RPEs ranging 2-10 with increments of 1. The UI should show a drop-down defaulting to Null, but then show values from a database lookup that offer the values as shown in "How it works" section below, perhaps with a drop-down picker.



#### How it works (2-10 Scale Example)

- **2:** Very light effort, like a slow walk.
- **3:** Light effort, comfortable pace, can easily talk.
- **4:** Light effort, comfortable pace, can easily talk.
- **5:** Moderate effort, breathing heavier, can still talk.
- **6:** Moderate effort, breathing heavier, can still talk.
- **7:** Hard effort, heavy breathing, difficult to talk.
- **8:** Hard effort, heavy breathing, difficult to talk.
- **9:** Very, very hard effort, almost all-out.
- **10:** Maximal, all-out effort. 



### image_006.jpg
<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_006.jpg" style="zoom: 25%;" />

Clicking the body (card?) of any WOD, Template, Movement in the system on any screen should display a view-only detail of that record using formatting from the markdown *if* no other action is currently assigned to that click event. Include either a back button, or a QuickLog link and Edit button right on the detailed view.


### image_003.jpg
<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_003.jpg" style="zoom: 25%;" />

The Workout Summary cards on the User Profile screen do not appear to show accurate information for all time windows. Verify and fix these numbers, particularly for All Time

### image_005.jpg

<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_005.jpg" style="zoom: 25%;" />

We need to add a Percent of Best section to the Stats screens. The section on this screen should go in this order: 1. Best Estimated 1RM, 2. Percent of Best, 3. Heaviest Lifts, 4. Performance Chart. all section retain their current content, but we need to add a section. All sections should become collapsible though, with Heaviest Lifts and Percent of Best defaulting to expanded.

The new Percent of Best section can appear just for Movements that record weights at this time. It should be displayed as a two column grid. Column 1 should be "% of Heaviest", and column two should be titled "lbs" or "kg" depending on the units chosen by the user settings.

"% of Heaviest" be the following values: 95%, 90%, 85%, 80%, 75%, 60%, 65%, 55%, 50%, 25%. The second column, "lbs" or "kg" should be equal to the percent of the very heaviest list completed with any number of reps.




### image_007.jpg
<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_007.jpg" style="zoom: 25%;" />

Same as image 006 above. Clicking the row should display a detailed view of the WOD, Workout Template, or Movement. Include either a back button, or a QuickLog link and Edit button right on the detailed view.


### image_004.jpg
<img src="/home/jcz/Github/actionlog/screenshots/12Jan2026/image_004.jpg" style="zoom: 25%;" />


Clicking a date in the calendar should open a detailed view of the Workout/WOD/Movement (all if more than one on that day) logged on that date. Include either a back button, or a QuickLog link and Edit button right on the detailed view.


