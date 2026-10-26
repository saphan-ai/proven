plugins {
    id("com.android.application") version "8.2.0"
}

dependencies {
    implementation("com.example:widget:1.4.0")
    implementation("com.example:quiet:1.9.0")
    implementation("com.example:variable:$widgetVersion")
}
