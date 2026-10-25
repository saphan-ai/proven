plugins {
    id("com.android.application") version "8.2.0"
}

dependencies {
    implementation("com.example:widget:1.2.0")
    implementation("com.example:quiet:2.0.0")
    implementation("com.example:variable:$widgetVersion")
}
